// Package userinput binds configuration struct fields to environment variables,
// command-line flags, and interactive terminal input.
//
// A configuration field participates in input processing when it has an input
// tag. The tag contains a comma-separated list of sources:
//
//	input:"env"           read an environment variable
//	input:"flag"          register a command-line flag
//	input:"term"          prompt on the terminal
//	input:"term,noecho"   request suppressed echo for string terminal input
//
// Sources may be combined. BindInput processes environment variables first,
// command-line flags second, and terminal input last. A command-line flag can
// therefore replace a value or validation result produced by an environment
// variable. Terminal input is requested only for fields not already recorded as
// supplied by an earlier source.
//
// The supported field kinds are bool, int, and string, including named types
// whose underlying kind is one of those three kinds.
//
// Additional struct tags control names, help text, and validation:
//
//	name:"host_name"              alternate input name
//	usage:"Server host name"      flag usage and terminal prompt text
//	min:"1" max:"65535"           integer range
//	regexp:"^[A-Za-z0-9.-]+$"     string validation
//
// Environment-variable names are uppercased and use underscores. Flag names are
// lowercased and use hyphens. SetPrefix applies a common prefix using the same
// transformations.
//
// Existing struct values serve as defaults. For terminal input, entering a blank
// line keeps the current value. Configuration errors such as invalid tags or an
// unsupported field kind cause a panic. BindInput reports input and validation
// errors. Per-field results recorded during binding are available through Supplied.
//
// A typical configuration is:
//
//	type Config struct {
//		Host     string `input:"env,flag,term" name:"host_name" usage:"Server host name" regexp:"^[A-Za-z0-9.-]+$"`
//		Port     int    `input:"env,flag,term" usage:"Server port number" min:"1" max:"65535"`
//		Password string `input:"term,noecho"   usage:"Administrator password"`
//	}
//
//	cfg := Config{Host: "localhost", Port: 8080}
//	ui := NewUserInput(&cfg)
//	ui.SetPrefix("myapp_")
//	errs := ui.BindInput()
//
package userinput
