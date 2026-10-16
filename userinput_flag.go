package userinput

import (
    "flag"
    "fmt"
    "os"
    "io"
    "sort"
    "errors"
    "strconv"
    "strings"
    r "reflect"
)


// NewFlagSet creates the FlagSet used by BindInput and returns it to the caller.
//
// The optional setname becomes the FlagSet name. The set uses flag.ContinueOnError.
// Call NewFlagSet after NewUserInput and before BindInput when the caller needs
// access to the FlagSet, for example to customize usage handling. It panics if a
// FlagSet has already been created for this UserInput.
func (ui *UserInput) NewFlagSet (setname ...string) *flag.FlagSet {

    if ui.flagSet != nil {
        panic("userinput.NewFlagSet: FlagSet already exists")
    }

    var sn string
    if len(setname) > 0 {
        sn = setname[0]
    }

    ui.flagSet = flag.NewFlagSet(sn, flag.ContinueOnError)

    ui.flagSet.SetOutput(io.Discard)
    ui.flagSet.Usage = func() {}

    return ui.flagSet
}


// PrintUsage prints help text with usage to stdout
func (ui *UserInput) PrintUsage () {

    var hasTerm bool

    // Sort by flagName
    var flags []*ConfiguredField
    var termFields []*ConfiguredField
    for _, cf := range ui.vogons {
        if cf.flagName != "" {
            flags = append(flags, cf)
        } else {
            if cf.termInput {
                hasTerm = true
                termFields = append(termFields, cf)
            }
        }
    }
    sort.Slice(flags, func(i, j int) bool {
        return flags[i].flagName < flags[j].flagName
    })

    fmt.Println()
    if ui.helpText != "" {
        fmt.Println(ui.helpText + "\n")
    }

    if len(flags) > 0 {
        fmt.Println("Options:")
    }

    spc := strings.Repeat(" ", 15)
    for _, cf := range flags {

        var typeName string
        var dfault string

        switch cf.fieldValue.Kind() {
        case r.Bool:
            typeName = "boolean"

        case r.Int:
            typeName = "integer"
            if !cf.fieldValue.IsZero() && !cf.suppressEcho {
                dfault = strconv.Itoa(int(cf.fieldValue.Int()))
            }

        case r.String:
            typeName = "string"
            if !cf.fieldValue.IsZero() && !cf.suppressEcho {
                dfault = cf.fieldValue.String()
            }
        }

        fmt.Printf("  -%s %s\n", cf.flagName, typeName)
        fmt.Printf("%s%s\n", spc, cf.usage)

        if cf.minSet || cf.maxSet {
            fmt.Printf("%sRange: %d to %d\n", spc, cf.min, cf.max)
        }

        if dfault != "" {
            fmt.Printf("%sDefault: %s\n", spc, dfault)
        }

        if cf.envName != "" {
            fmt.Printf("%sEnv var: %s\n", spc, cf.envName)
        }

        if cf.termInput {
            hasTerm = true
            var echo string
            if cf.suppressEcho {
                echo = ", echo suppressed"
            }
            fmt.Printf("%sTerminal: prompted%s\n", spc,  echo)
        }

    }

    fmt.Printf("  -h, -help\n%sDisplay this help\n", spc)


    if len(termFields) > 0 {
        //var dfault string

        fmt.Println("\nTerminal Input:")

        for _, cf := range termFields {
            var suppressed string
            if cf.suppressEcho {
                suppressed = " (echo suppressed)"
            }
            fmt.Printf("    %s%s\n", cf.usage, suppressed)
        }
    }

    if hasTerm {
        fmt.Println("\nNote: Terminal input is requested only when the field has not been\n"+
        "      supplied by another input source.")
    }
}

// bindFlagSet registers configured flags, parses os.Args, validates explicitly
// supplied values, and commits valid values to their destination fields. Parsed
// values are staged first so an invalid flag does not overwrite the current field.
func (ui *UserInput) bindFlagSet () []error {

    var errs []error

    if ui.flagSet == nil {
        ui.NewFlagSet()
    }

    //ui.flagSet.Usage = ui.PrintUsage

    flagToField := make(map[string]string)

    for fieldName, cf := range ui.vogons {

        if cf.flagName == "" {
            continue
        }

        flagToField[cf.flagName] = fieldName

        switch cf.fieldValue.Kind() {
        case r.Bool:
            ui.flagSet.Var(
                &boolValue{field: cf.stagedValue},
                cf.flagName,
                cf.usage)

        case r.Int:
            ui.flagSet.Var(
                &intValue{field: cf.stagedValue},
                cf.flagName,
                cf.usage)

        case r.String:
            ui.flagSet.Var(
                &stringValue{field: cf.stagedValue},
                cf.flagName,
                cf.usage)
        }
    }

    if err := ui.flagSet.Parse(os.Args[1:]); err != nil {
        if errors.Is(err, flag.ErrHelp) {
            ui.helpRequested = true
            return nil
        }
        return []error{err}
    }

    // Visit validates only flags explicitly supplied on the command line.
    ui.flagSet.Visit(func(f *flag.Flag) {

        fieldName, ok := flagToField[f.Name]
        if !ok {
            return
        }

        cf := ui.vogons[fieldName]

        switch cf.fieldValue.Kind() {
        case r.Bool:
            // do nothing

        case r.Int:
            v := cf.stagedValue.Int()
    		if int(v) < cf.min || int(v) > cf.max {
                errs = append(errs, fmt.Errorf("value out of range (%d), for flag -%s, expecting %d to %d", v, cf.flagName, cf.min, cf.max))
                return
            }

        case r.String:
            s := cf.stagedValue.String()
            if cf.re != nil && !cf.re.MatchString(s) {
                errs = append(errs, fmt.Errorf(`invalid string "%s", for flag -%s`, s, cf.flagName))
                return
            }

        }

        ui.supplied[fieldName] = true

    })

    return errs
}
