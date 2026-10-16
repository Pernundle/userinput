# pernundle/userinput

`userinput` binds fields in a Go configuration struct to three common sources of program input:

- environment variables
- command-line flags
- interactive terminal prompts

The package lets one struct describe where input may come from, how it is named, how it is validated, and what default value should be retained when the user supplies nothing.

Supported field kinds are `bool`, `int`, and `string`, including named types whose underlying kind is one of those three.

## Example

```go
type Config struct {
    Host       string `input:"env,flag,term" name:"host_name" usage:"Server host name" regexp:"^[A-Za-z0-9.-]+$"`
    Port       int    `input:"env,flag,term" usage:"Server port number" min:"1" max:"65535"`
    Enabled    bool   `input:"env,flag,term" usage:"Enable server"`
    AdminAcct  string `input:"term" usage:"Administrator account"`
    Password   string `input:"term,noecho" usage:"Administrator password"`
}

cfg := Config{
    Host: "localhost",
    Port: 8080,
}

ui := userinput.NewUserInput(&cfg)
ui.SetPrefix("myapp_")

if errs := ui.BindInput(); len(errs) != 0 {
    for _, err := range errs {
        fmt.Println(err)
    }
    return
}
```

With the prefix above, `Host` is available as:

```text
Environment: MYAPP_HOST_NAME
Flag:        -myapp-host-name
Terminal:    Server host name
```

If no source supplies a value, the existing value in `cfg` remains unchanged.

## Input Sources

The `input` struct tag selects one or more sources for a field.

| Tag | Meaning |
| --- | --- |
| `input:"env"` | Read from an environment variable. |
| `input:"flag"` | Register a command-line flag. |
| `input:"term"` | Prompt on the terminal. |
| `input:"env,flag"` | Accept environment or flag input. |
| `input:"env,term"` | Accept environment or terminal input. |
| `input:"flag,term"` | Accept flag or terminal input. |
| `input:"env,flag,term"` | Accept all three sources. |
| `input:"term,noecho"` | Prompt without echoing string or integer input. |
| `input:"flag,noecho"` | Hide the default value from package-generated flag help. |

Do not place spaces between items in an `input` tag.

`noecho` is not valid for `bool` fields.

## Input Order and Precedence

`BindInput` processes sources in this order:

1. environment variables
2. command-line flags
3. terminal input

When both an environment variable and a command-line flag supply a valid value for the same field, the flag value takes precedence.

Terminal input is requested only when that field has not already been recorded as supplied by an earlier source. This makes terminal input useful as an interactive fallback rather than an additional mandatory step.

Errors encountered during environment variable and flag parsing will be output to the terminal.


## Struct Tags

### `input`

Selects the permitted input sources and optional `noecho` behavior.

```go
Host     string `input:"env,flag,term"`
Password string `input:"term,noecho"`
```

At least one exported field must have an `input` tag.

### `name`

Overrides the field name used to generate environment-variable and flag names.

```go
Host string `input:"env,flag" name:"server_host"`
```

This produces:

```text
Environment: SERVER_HOST
Flag:        -server-host
```

Names may contain letters, digits, underscores, and hyphens. A name must begin with a letter, end with a letter or digit, and may not exceed 100 characters.

The exported `NameRegexp` constant contains the regular expression used for this validation.

### `usage`

Provides help text for command-line flags and prompt text for terminal input.

```go
Port int `input:"flag,term" usage:"Server port number"`
```

When `usage` is omitted:

- flag help uses the Go field name
- terminal input uses `Enter <FieldName>`

### `regexp`

Validates string input.

```go
Host string `input:"env,flag,term" regexp:"^[A-Za-z0-9.-]+$"`
```

`regexp` may be used only with string fields. An empty or malformed regular expression is considered a configuration error and will cause a panic.

### `min` and `max`

Set an inclusive range for integer input.

```go
Port int `input:"env,flag,term" min:"1" max:"65535"`
```

Either tag may be used independently. When omitted, the corresponding limit is the platform `int` minimum or maximum.

`min` and `max` may be used only with integer fields, and `min` may not be greater than `max`.


### Prefixes

`SetPrefix` prepends a common prefix to every configured environment-variable and flag name.

```go
ui.SetPrefix("myapp_")
```

A base name of `host_name` then becomes:

```text
MYAPP_HOST_NAME
-myapp-host-name
```

The prefix is transformed using the same environment and flag naming rules. `SetPrefix` must be called before `BindInput`, may be called only once, and may not be blank.


## Default Values

Existing struct values are defaults.

```go
cfg := Config{
    Host: "localhost",
    Port: 8080,
}
```

If a field is not supplied by any configured source, its current value remains unchanged.

For terminal input, a blank response keeps the current value.

Nonzero string and integer defaults are displayed in terminal prompts unless `noecho` is set. Boolean prompts show the current default as `[Y/n]` or `[y/N]`.

## Environment Variables

Missing environment variables are ignored.

### Strings

Strings may be checked with the `regexp` tag. An empty environment string is accepted as a supplied value and bypasses regular-expression validation.

### Integers

Integer values must parse as a Go `int` and satisfy any `min` or `max` constraints.

### Booleans

Environment and flag booleans accept these values, case-insensitive:

```text
0
1
t
f
true
false
```

## Command-Line Flags

String and integer values are validated using the same `regexp`, `min`, and `max` configuration used by the other input sources.

Boolean fields behave as Go boolean flags and may therefore be written in the usual forms, including a bare true flag such as:

```text
-enabled
```

The package installs `PrintUsage` as the `FlagSet` usage function. Generated help can include:

- flag name and type
- usage text
- integer range
- nonzero string or integer default
- associated environment variable
- whether terminal input is available
- whether terminal echo is suppressed

`noecho` suppresses display of string and integer defaults in this generated help.

### Custom FlagSet

Call `NewFlagSet` after `NewUserInput` and before `BindInput` when the caller needs direct access to the `flag.FlagSet`.

```go
ui := userinput.NewUserInput(&cfg)
fs := ui.NewFlagSet("mycommand")
_ = fs
```

Calling `NewFlagSet` more than once for the same `UserInput` causes a panic.

## Terminal Input

Interactive input is read directly from `/dev/tty`, rather than from standard input. This allows terminal prompting even when stdin is redirected.

String and integer fields repeat the prompt after invalid input.

Boolean terminal input accepts only "Y" or "N", case insensitive.

A blank response preserves the current default.

### Suppressed Echo

Add `noecho` to a string or integer field when typed input should not be displayed:

```go
Password string `input:"term,noecho" usage:"Administrator password"`
PIN      int    `input:"term,noecho" usage:"Administrator PIN"`
```

A field using `noecho` does not display its default value in the prompt or package-generated flag help.

`noecho` cannot be used with a boolean field.

## Validation and Errors

The package distinguishes configuration mistakes from user-input errors.

### Configuration errors

Programming or struct-tag mistakes cause a panic during setup. Examples include:

- `cfg` is not a non-nil pointer to a struct
- no exported field exists
- no field has an `input` tag
- a configured field has an unsupported kind
- an input source name is unknown
- a `name` tag is invalid
- duplicate generated names are found
- `regexp` is empty, malformed, or applied to a non-string field
- `min` or `max` is malformed or applied to a non-integer field
- `min` is greater than `max`
- `noecho` is applied to a boolean field
- `SetPrefix` is called with an invalid prefix or called more than once
- `BindInput` is called more than once

### User-input errors

Input and validation failures are returned by `BindInput` as `[]error`.

```go
errs := ui.BindInput()
if len(errs) != 0 {
    for _, err := range errs {
        log.Print(err)
    }
}
```

Terminal validation is interactive: invalid values are reported at the terminal and the package prompts again.

## Supplied Values

`Supplied` returns a copy of the package's internal map of fields for which an input source recorded a value or validation result.

```go
results := ui.Supplied()

if result, ok := results["Port"]; ok {
    if result.Valid {
        fmt.Println("Port was supplied and validated")
    } else {
        fmt.Println(result.Err)
    }
}
```

The map is keyed by Go struct field name, not by the environment-variable or flag name.

```go
type SuppliedResult struct {
    Valid bool
    Err   error
}
```


## Public API

### Construction and binding

```go
func NewUserInput(cfg any) *UserInput
func (ui *UserInput) BindInput() []error
```

### Naming

```go
func (ui *UserInput) SetPrefix(s string)
```

### Supplied-input reporting

```go
func (ui *UserInput) Supplied() map[string]SuppliedResult
```

### Flag handling

```go
func (ui *UserInput) NewFlagSet(setname ...string) *flag.FlagSet
func (ui *UserInput) PrintUsage()
```

### Terminal helpers

The userinput package provides two helper functions for terminal input.

```go
func TermReadInput(prompt string) (string, error)
func TermReadPassword(prompt string) (string, error)
```

`TermReadPassword` reads text directly from `/dev/tty` and returns it as a string.

`TermReadPassword` reads text directly from `/dev/tty` without echoing it and returns it as a string.

### Exported supporting types and constants

```go
type InputMode int

const (
    NormalEcho InputMode = iota
    SuppressEcho
)

const NameRegexp = `^[A-Za-z](?:[A-Za-z0-9_-]{0,98}[A-Za-z0-9])?$`

type SuppliedResult struct {
    Valid bool
    Err   error
}

type SuppliedMap map[string]SuppliedResult
```

## Platform Note

Interactive terminal support currently depends on `/dev/tty` and `golang.org/x/term`. The terminal portion of the package is therefore written for Unix-like systems that provide `/dev/tty`.

Environment-variable and command-line flag binding do not themselves require interactive terminal input.

**Note:** *Documentation written by Artificial Intelligence and edited by Alleged Intelligence.*
