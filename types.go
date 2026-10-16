package userinput

import (
	"flag"
	r "reflect"
	"regexp"
)

func init() {
	compiledNameRegexp = regexp.MustCompile(NameRegexp)
	compiledTrueFalseRe = regexp.MustCompile(NameRegexp)
}

var compiledNameRegexp  *regexp.Regexp
var compiledTrueFalseRe *regexp.Regexp

// InputMode controls whether terminal input is echoed while it is entered.
type InputMode int

const (
	// NormalEcho leaves normal terminal echo enabled.
	NormalEcho InputMode = iota

	// SuppressEcho disables terminal echo while a value is read.
	SuppressEcho
)


// Mimic the flag package error handling... somewhat.
type BindOption int
const ContinueOnError  BindOption = 1   // Call os.Exit(2) or for -h/-help Exit(0)



// NameRegexp is the regular expression used to validate flag and env names.
// It allows names containing letters, digits, hyphens, and underscores.
// Names must begin with a letter and end with a letter or number.
// May not exceed 100 characters.
const   NameRegexp = `^[A-Za-z](?:[A-Za-z0-9_-]{0,98}[A-Za-z0-9])?$`

// SuppliedMap maps configuration struct field names to their supplied-input results.
type SuppliedMap map[string]bool

// UserInput coordinates binding and validation for a configuration struct.
//
// A UserInput is created with NewUserInput. Its configuration is inspected once
// at construction time, after which SetPrefix or NewFlagSet may be called before
// BindInput performs the actual input processing.
type UserInput struct {

    // provides help text prepended to the generated option list
    helpText        string

    // will be true if help was requested by user
    helpRequested   bool

	// reflection value of the supplied config struct
	cfgValue        r.Value

	// prefix is optionally prepended to every environment variable and flag name.
	//
	// With prefix set to "MYAPP_", an env tag of "HOST" results in the environment
	// variable name "MYAPP_HOST" and a command-line flag name "myapp-host"
	prefix          string

	// The following fields record conditions encountered while examining cfg.
	// They are used after processing to distinguish configuration mistakes
	// from ordinary cases where an individual environment variable is absent.
	hasExported      bool
	hasEnv           bool
	hasFlag          bool
	hasTag           bool
	hasSupportedKind bool
	inputBound       bool

	// contains the fully inspected and administratively approved fields.
	vogons          map[string]*ConfiguredField

	// contains fields that were actually input by the user and
	// successfully bound to struct fields.
	supplied        SuppliedMap

	// contains input validation errors encountered while binding values.
	// Unlike programmer configuration errors, these are returned to the caller.
	errs            []error

	//
	flagSet         *flag.FlagSet
}

// ConfiguredField contains the inspected metadata and reflected destination for
// one configuration field. Its fields are intentionally internal to the package.
type ConfiguredField struct {
	fieldName   string
	flagName    string
	envName     string
	usage       string
	prompt      string

	termInput    bool
	suppressEcho bool

	min         int
	max         int

    minSet      bool
    maxSet      bool

	re          *regexp.Regexp

	fieldKind   r.Kind
	fieldInfo   r.StructField
	fieldValue  r.Value // destination in cfg

	stagedValue r.Value // parsed, validated value waiting to be committed
}
