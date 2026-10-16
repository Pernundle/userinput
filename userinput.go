package userinput

import (
    "os"
	"fmt"
	r "reflect"
	"regexp"
	"strings"
)

// NewUserInput creates a UserInput bound to cfg.
//
// cfg must be a non-nil pointer to a struct. The struct is inspected immediately,
// and invalid binding configuration causes NewUserInput to panic. Supported field
// kinds are bool, int, and string.
func NewUserInput (cfg any) *UserInput {

	// Reflection requires cfg to be a pointer so that field values can be
	// modified directly.
	v := r.ValueOf(cfg)

	if v.Kind() != r.Ptr || v.IsNil() {
		panic("cfg must be a non-nil pointer to a struct")
	}

	v = v.Elem()

	if v.Kind() != r.Struct {
		panic("cfg must be a pointer to a struct")
	}

	ui := &UserInput{
		cfgValue: v,
		supplied: make(SuppliedMap),
		vogons:   make(map[string]*ConfiguredField),
		errs:     []error{},
	}

	// inspectCfg does not return errors. It panics if there are any problems.
	ui.inspectCfg()

	return ui
}

// BindInput binds configured fields from environment variables, command-line
// flags, and terminal input, in that order.
//
// Command-line flags may replace values or validation results produced by
// environment variables. Terminal prompts are used only for fields that have not
// already been recorded as supplied. Existing struct values remain as defaults
// when no source supplies a replacement value.
//
// BindInput returns user-input and validation errors. It panics if called more
// than once on the same UserInput.
func (ui *UserInput) BindInput () []error {

	if ui.inputBound {
		panic("BindInput has already been called")
	}

    hasError := false

	if errs := ui.bindEnv(); errs != nil {
        for _, err := range errs {
            fmt.Printf("%s\n", err.Error())
        }
        hasError = true
	}

	if errs := ui.bindFlagSet(); errs != nil {
        for _, err := range errs {
            fmt.Printf("%s\n", err.Error())
        }
        hasError = true
	}

    if ui.helpRequested {
        os.Exit(0)
    }

    if hasError {
        ui.PrintUsage()
        os.Exit(2)
    }

	if err := ui.bindTerminal(); err != nil {
		return []error{err}
	}

    ui.commitStagedValues()

	ui.inputBound = true
	return nil
}


func (ui *UserInput) commitStagedValues () {
	for f, _ := range ui.supplied {
        if !ui.vogons[f].fieldValue.CanSet() {
            continue
        }
        ui.vogons[f].fieldValue.Set(ui.vogons[f].stagedValue)
    }
}


// inspectCfg examines cfg and builds the metadata used by each input source.
func (ui *UserInput) inspectCfg () {

	v := ui.cfgValue

	// seen is used to check for duplicate aliases
	seen := make(map[string]bool)

	// Examine each field, collecting the ones which are eligible for binding.
	for i := 0; i < v.NumField(); i++ {

		fieldInfo := v.Type().Field(i)

		// Ignore unexported fields
		if !fieldInfo.IsExported() {
			continue
		}

		ui.hasExported = true

		// initialize the configured field struct
		cf := ConfiguredField{
			fieldName:  fieldInfo.Name,
			fieldInfo:  fieldInfo,
			fieldValue: v.Field(i),
			fieldKind:  v.Field(i).Kind(),
			stagedValue: r.New(v.Field(i).Type()).Elem(),
		}

		if !supportedKind(cf.fieldValue) {
			panic(fmt.Sprintf("Kind not supported: %v", cf.fieldKind))
		}
		//ui.hasSupportedKind = true

		alias := getName(fieldInfo)
		alias = ui.prefix + alias

		// Panic if duplicate alias is found
		if seen[alias] {
			panic(fmt.Sprintf("Duplicate name: %s at field %s", alias, fieldInfo.Name))
		}
		seen[alias] = true

		// check for input tag
		input, ok := cf.fieldInfo.Tag.Lookup("input")
		if !ok {
			continue
		}
		for _, inp := range strings.Split(input, ",") {
			switch inp {
			case "env":
				cf.envName = strings.ReplaceAll(strings.ToUpper(alias), "-", "_")
			case "flag":
				cf.flagName = strings.ReplaceAll(strings.ToLower(alias), "_", "-")
			case "term":
				cf.termInput = true
			case "noecho":
				if cf.fieldKind == r.Bool {
					panic(fmt.Sprintf("cannot use 'noecho' with type bool: %s", cf.fieldName))
				}
				cf.suppressEcho = true
			default:
				panic(fmt.Sprintf("undefined input source: %s in field %s", inp, cf.fieldName))
			}
		}

		ui.hasTag = true
		cf.usage = getUsage(cf.fieldInfo)
        if cf.termInput {
            cf.prompt = getPrompt(cf.fieldInfo)
        }

		if regx, ok := cf.fieldInfo.Tag.Lookup("regexp"); ok {
			if regx == "" {
				panic(fmt.Sprintf("tag regexp cannot be empty: %s", cf.fieldName))
			}
			if cf.fieldKind != r.String {
				panic(fmt.Sprintf("tag regexp can only be used with type string: %s", cf.fieldName))
			}
			// if regexp is malformed, it will panic here
			cf.re = regexp.MustCompile(regx)
		}

		if cf.fieldKind == r.Int {
			// getMinMax will panic if min or max values are invalid
			getMinMax(&cf)
		} else {
			if _, ok := cf.fieldInfo.Tag.Lookup("min"); ok {
				panic(fmt.Sprintf("tag min can only be used with type int: %s", cf.fieldName))
			}
			if _, ok := cf.fieldInfo.Tag.Lookup("max"); ok {
				panic(fmt.Sprintf("tag max can only be used with type int: %s", cf.fieldName))
			}
		}

		// Apply the field information to the vogon map
		ui.vogons[cf.fieldName] = &cf

	}

	// Any of these conditions indicate an invalid binding configuration.
	if !ui.hasExported {
		panic("No exported field was found in cfg")
	}
	//if !ui.hasSupportedKind {
	//    panic("No field with a supported type was found in cfg")
	//}
	if !ui.hasTag {
		panic("No field was found with an `input` tag")
	}

}


// SetPrefix prepends s to every configured environment-variable and flag name.
//
// For environment variables, the prefix is uppercased and hyphens become
// underscores. For flags, it is lowercased and underscores become hyphens. The
// resulting names must satisfy NameRegexp.
//
// SetPrefix must be called before BindInput. It panics if s is blank, if a prefix
// has already been set, or if a resulting name is invalid.
func (ui *UserInput) SetPrefix (s string) {
	if s == "" {
		panic("prefix cannot be blank")
	}
	if ui.prefix != "" {
		panic("prefix already set")
	}

	ui.prefix = s
	envPrefix := strings.ReplaceAll(strings.ToUpper(s), "-", "_")
	flagPrefix := strings.ReplaceAll(strings.ToLower(s), "_", "-")

	// validate prefix + name
	for fieldName, cf := range ui.vogons {
		if cf.envName != "" {
			if !compiledNameRegexp.MatchString(envPrefix + cf.envName) {
				panic(fmt.Sprintf("invalid prefix/name: %s", fieldName))
			}
		}
		if cf.flagName != "" {
			if !compiledNameRegexp.MatchString(flagPrefix + cf.flagName) {
				panic(fmt.Sprintf("invalid prefix/name: %s", fieldName))
			}
		}
	}

	// apply changes
	for _, cf := range ui.vogons {
		if cf.envName != "" {
			cf.envName = envPrefix + cf.envName
		}
		if cf.flagName != "" {
			cf.flagName = flagPrefix + cf.flagName
		}
	}

}

// Supplied returns a copy of the results recorded for supplied fields.
//
// The map is keyed by configuration struct field name, not by environment-variable
// or flag name. Valid is true when the supplied value was validated and assigned.
// When validation fails, Valid is false and Err contains the error. Fields for
// which no input source recorded a value are absent from the map.
func (ui *UserInput) Supplied () SuppliedMap {
	m := make(SuppliedMap, len(ui.supplied))
	for name, _ := range ui.supplied {
		m[name] = true
	}
	return m
}


func (ui *UserInput) SetHelpText (txt string) {
    ui.helpText = txt
}

func (ui *UserInput) HelpRequested() bool {
    return ui.helpRequested
}

// Documentation written by Artificial Intelligence and edited by Alleged Intelligence.
