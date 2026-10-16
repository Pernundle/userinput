package userinput

import (
    "os"
    "fmt"
  r "reflect"
    "testing"
  . "github.com/smartystreets/goconvey/convey"
)



func Test010_EnvInput (t *testing.T) {

    // Set environment variables
    t.Setenv("BNAME",   "Input B")
    t.Setenv("C",       "10")
    t.Setenv("D",       "true")

    Convey("bindEnv\n", t, func() {

        type cfgEnvInput struct {
            B   string      `input:"env"  name:"bname"  regexp:"^[a-zA-z0-9 ]+$"`
            C   int         `input:"env"  min:"10"  max:"100"`
            D   bool        `input:"env"`
        }

        var cfg cfgEnvInput
        ui := NewUserInput(&cfg)

    	errs := ui.bindEnv()
        So(errs, ShouldBeNil)
        ui.commitStagedValues()

        So(len(ui.vogons), ShouldEqual, 3)

        cf, ok := ui.vogons["B"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "B")
        So(cf.envName,          ShouldEqual, "BNAME")
        So(cfg.B, ShouldEqual, "Input B")

        cf, ok = ui.vogons["C"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "C")
        So(cf.envName,          ShouldEqual, "C")
        So(cfg.C, ShouldEqual,  10)

        cf, ok = ui.vogons["D"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "D")
        So(cf.envName,          ShouldEqual, "D")
        So(cfg.D, ShouldEqual,  true)

    })

}


func Test020_FlagInput (t *testing.T) {

    // Setup command line flag
    osArgs := os.Args
    defer func() {
        os.Args = osArgs
    }()
    os.Args = []string{
        "mycommand",
        "-bname=Binput",
        "-c=11",
        "-d=false",
    }

    Convey("bindFlagSet\n", t, func() {

        type cfgFlagInput struct {
            B   string      `input:"flag"  name:"bname"  regexp:"^[a-zA-z0-9 ]+$"`
            C   int         `input:"flag"  min:"10"  max:"100"`
            D   bool        `input:"flag"`
        }

        var cfg cfgFlagInput
        ui := NewUserInput(&cfg)

    	errs := ui.bindFlagSet()
        So(errs, ShouldBeNil)

        ui.commitStagedValues()

        cf, ok := ui.vogons["B"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "B")
        So(cf.flagName,          ShouldEqual, "bname")
        So(cfg.B, ShouldEqual, "Binput")

        cf, ok = ui.vogons["C"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "C")
        So(cf.flagName,          ShouldEqual, "c")
        So(cfg.C, ShouldEqual,  11)

        cf, ok = ui.vogons["D"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "D")
        So(cf.flagName,          ShouldEqual, "d")
        So(cfg.D, ShouldEqual,  false)

    })

}



func Test030_BindTerminal (t *testing.T) {

    originalFunc := ttyReadLine
    defer func() {
        ttyReadLine = originalFunc
    }()

    ttyReadLine = func(prompt string, mode InputMode) (string, error) {
        fmt.Printf("\nprompt: (%s)\n", prompt)
        switch prompt {
            case "Enter B":
                return "Terminal Input", nil
            case "Enter C":
                return "42", nil
            case "Enter D [y/N]":
                return "y", nil
            default:
                return "", fmt.Errorf("Unknown prompt")
        }
    }

    Convey("bindTerminal\n", t, func() {

        type cfgEnvFlagInput struct {
            B   string      `input:"term"`
            C   int         `input:"term"`
            D   bool        `input:"term"`
        }

        var cfg cfgEnvFlagInput
        ui := NewUserInput(&cfg)

    	err := ui.bindTerminal()
        So(err, ShouldBeNil)

        ui.commitStagedValues()

        So(cfg.B, ShouldEqual, "Terminal Input")
        So(cfg.C, ShouldEqual,  42)
        So(cfg.D, ShouldEqual,  true)

    })

}


func Test040_BindTerminalWithDefaultValues (t *testing.T) {

    originalFunc := ttyReadLine
    defer func() {
        ttyReadLine = originalFunc
    }()

    // return empty string
    ttyReadLine = func(prompt string, mode InputMode) (string, error) {
        fmt.Printf("\nprompt: (%s)\n", prompt)
        return "", nil
    }

    Convey("bindTerminal\n", t, func() {

        type cfgEnvFlagInput struct {
            B   string      `input:"term"`
            C   int         `input:"term"`
            D   bool        `input:"term"`
        }

        cfg := cfgEnvFlagInput{
            B:      "Default",
            C:      42,
            D:      true,
        }

        ui := NewUserInput(&cfg)

    	err := ui.bindTerminal()
        So(err, ShouldBeNil)

        ui.commitStagedValues()

        So(cfg.B, ShouldEqual, "Default")
        So(cfg.C, ShouldEqual,  42)
        So(cfg.D, ShouldEqual,  true)

    })

}


func Test050_FlagOverride (t *testing.T) {

    // Set environment variables
    t.Setenv("BNAME",   "Input B")
    t.Setenv("C",       "10")
    t.Setenv("D",       "true")

    // Setup command line flags
    osArgs := os.Args
    defer func() {
        os.Args = osArgs
    }()
    os.Args = []string{
        "mycommand",
        "-b=Binput",
        "-c=11",
        "-d=false",
    }

    Convey("Flag values will override env values\n", t, func() {

        type cfgEnvFlagInput struct {
            B   string      `input:"env,flag"`
            C   int         `input:"env,flag"`
            D   bool        `input:"env,flag"`
        }

        var cfg cfgEnvFlagInput
        ui := NewUserInput(&cfg)

    	errs := ui.bindEnv()
        So(errs, ShouldBeNil)

    	errs = ui.bindFlagSet()
        So(errs, ShouldBeNil)

        ui.commitStagedValues()

        cf, ok := ui.vogons["B"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "B")
        So(cf.flagName,          ShouldEqual, "b")
        So(cfg.B, ShouldEqual, "Binput")

        cf, ok = ui.vogons["C"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "C")
        So(cf.flagName,          ShouldEqual, "c")
        So(cfg.C, ShouldEqual,  11)

        cf, ok = ui.vogons["D"]
        So(ok, ShouldBeTrue)
        So(cf.fieldName,        ShouldEqual, "D")
        So(cf.flagName,          ShouldEqual, "d")
        So(cfg.D, ShouldEqual,  false)

    })

}


func Test060_TerminalNotCalled (t *testing.T) {

    // Set environment variables
    t.Setenv("BNAME",   "Input B")
    t.Setenv("C",       "10")


    // Setup command line flags
    osArgs := os.Args
    defer func() {
        os.Args = osArgs
    }()
    os.Args = []string{
        "mycommand",
        "-d=false",
    }

    originalFunc := ttyReadLine
    defer func() {
        ttyReadLine = originalFunc
    }()
    ttyReadLine   = func(prompt string, mode InputMode) (string, error) {
        fmt.Printf("\n*** prompt: (%s)\n", prompt)
        return "", fmt.Errorf("Terminal should not be called")
    }

    Convey("Terminal input not called when env or flag input\n", t, func() {

        // Terminal input should not be called as input has already been provided

        type cfgEnvFlagInput struct {
            B   string      `input:"env,flag,term"  name:"bname"  regexp:"^[a-zA-z0-9 ]+$"`
            C   int         `input:"env,term"  min:"10"  max:"100"`
            D   bool        `input:"flag,term"`
        }

        var cfg cfgEnvFlagInput
        ui := NewUserInput(&cfg)

    	errs := ui.bindEnv()
        So(errs, ShouldBeNil)

    	errs = ui.bindFlagSet()
        So(errs, ShouldBeNil)

    	err := ui.bindTerminal()
        So(err, ShouldBeNil)

        ui.commitStagedValues()

        So(cfg.B, ShouldEqual, "Input B")
        So(cfg.C, ShouldEqual,  10)
        So(cfg.D, ShouldEqual,  false)

    })

}




func Test070_BindTerminalNoEcho (t *testing.T) {

    originalFunc := ttyReadLine
    defer func() {
        ttyReadLine = originalFunc
    }()

    ttyReadLine = func(prompt string, mode InputMode) (string, error) {
        fmt.Printf("\nprompt: (%s)\n", prompt)
        switch prompt {
            case "Enter B":
                return "Terminal Input", nil
            case "Enter C":
                return "42", nil
            default:
                return "", fmt.Errorf("Unknown prompt")
        }
    }

    Convey("bindTerminal\n", t, func() {

        // Terminal input will be called as neither env nor flag are provided

        type cfgEnvFlagInput struct {
            B   string      `input:"term,noecho"`
            C   int         `input:"term,noecho"`
        }

        cfg := cfgEnvFlagInput{
            B:      "Default",
            C:      7334,
        }

        ui := NewUserInput(&cfg)

    	err := ui.bindTerminal()
        So(err, ShouldBeNil)

        ui.commitStagedValues()

        So(cfg.B, ShouldEqual, "Terminal Input")
        So(cfg.C, ShouldEqual,  42)

    })

}

func Test080_NamedTypes (t *testing.T) {

    type  myBool    bool
    type  myString  string
    type  myInt     int

    type cfgNamedTypes struct {
        A           myString  `input:"flag,env"`
        B           myInt     `input:"flag,env"`
        C           myBool    `input:"flag,env"`
    }

    Convey("Prove Named Types are Honored\n", t, func() {

        var cfg cfgNamedTypes
        ui := NewUserInput(&cfg)

        So(len(ui.vogons), ShouldEqual, 3)

        cf, ok := ui.vogons["A"]
        So(ok, ShouldBeTrue)
        So(cf.fieldKind,        ShouldEqual, r.String)

        cf, ok = ui.vogons["B"]
        So(ok, ShouldBeTrue)
        So(cf.fieldKind,        ShouldEqual, r.Int)

        cf, ok = ui.vogons["C"]
        So(ok, ShouldBeTrue)
        So(cf.fieldKind,        ShouldEqual, r.Bool)

    })

}

func Test090_DefaultValues (t *testing.T) {

    type cfgFlag struct {
        A   string      `input:"flag"`
        B   int         `input:"flag"`
        C   bool        `input:"flag"`
    }

    Convey("Prove Default Values Remain\n", t, func() {

        cfg := cfgFlag{
            A:      "new string",
            B:      1320,
            C:      true,
        }

        // Clear command line
        osArgs := os.Args
        defer func() {
            os.Args = osArgs
        }()
        os.Args = []string{
            "mycommand",
        }

        ui := NewUserInput(&cfg)
        ui.BindInput()

        // Prove that binding input with no flags did not change default values
        So(cfg.A, ShouldEqual, "new string")
        So(cfg.B, ShouldEqual, 1320)
        So(cfg.C, ShouldEqual, true)

    })

}



// // // // // // // // // // // // // // //

func Test100_BindInput_AllMethods (t *testing.T) {

    // Set one environment variable
    t.Setenv("BNAME",   "Input B")

    // Setup one command line flag
    osArgs := os.Args
    defer func() {
        os.Args = osArgs
    }()
    os.Args = []string{
        "mycommand",
        "-c=99",
    }

    originalFunc := ttyReadLine
    defer func() {
        ttyReadLine = originalFunc
    }()
    ttyReadLine   = func(prompt string, mode InputMode) (string, error) {
        fmt.Printf("\n*** prompt: (%s)\n", prompt)
        if prompt == "Enter D [y/N]" {
            return "y", nil
        }
        return "", fmt.Errorf("Terminal should not be called")
    }

    Convey("bindEnv, bindFlagSet, and bindTerminal\n", t, func() {

        // Terminal input should not be called as input has already been provided

        type cfgEnvFlagInput struct {
            B   string      `input:"env,flag,term"  name:"bname"  regexp:"^[a-zA-z0-9 ]+$"`
            C   int         `input:"env,flag,term"  min:"10"  max:"100"`
            D   bool        `input:"env,flag,term"`
        }

        var cfg cfgEnvFlagInput
        ui := NewUserInput(&cfg)

    	errs := ui.BindInput()
        So(errs, ShouldBeNil)

        So(cfg.B, ShouldEqual, "Input B")
        So(cfg.C, ShouldEqual,  99)
        So(cfg.D, ShouldEqual,  true)

        m := ui.Supplied()
        So( len(m), ShouldEqual, 3 )

        So(m["B"], ShouldBeTrue)
        So(m["C"], ShouldBeTrue)
        So(m["D"], ShouldBeTrue)

    })

}





/***  ***  ***  ***  ***  ***  ***  ***  ***  ***  ***  ***

Positive tests to prove package features.

Input sources:

    input:"env"
    input:"flag"
    input:"term"
    input:"env,flag"
    input:"env,flag,term"
    input:"env,term"
    input:"flag,term"
    input:"flag"
    input:"term,noecho"
    (test noecho with string and int)


Additional tags:

    name:"new-name"
    usage:"flag usage or terminal prompt"
    regexp:"properly formed regular expression"
    min:"integer"
    max:"integer"


Naming:
    default generated names
    name override
    prefix
    prefix + name
    env capitalization
    flag hyphenation


Data types:

    int
    string
    bool
    named types based on int, string, or bool.


Defaults:

    Prove that an existing struct value remains untouched
    when its corresponding input is not supplied.


Supplied Values:

    successfully supplied env field appears
    successfully supplied flag field appears
    successfully supplied terminal field appears
    unsupplied field does not appear
    flag or env input prevents terminal input.


Input precedence:

    (default < env < flag)
    (default < terminal)


Methods:

    func NewUserInput(cfg any) *UserInput
    func (ui *UserInput) BindInput() []error
    func (ui *UserInput) SetPrefix(s string)
        Prove prefix affect env and flag names appropriately.
        Prove prefix affects name tags as well as field names.

    func (ui *UserInput) Supplied() map[string]SuppliedResult
        Prove Supplied() returns a copy of the map


/***  ***  ***  ***  ***  ***  ***  ***  ***  ***  ***  ***/


/*
func Test110_Help_flagSet (t *testing.T) {

    // Setup command line
    osArgs := os.Args
    defer func() {
        os.Args = osArgs
    }()
    os.Args = []string{
        "mycommand",
        "-h",
    }


    Convey("Show help, all input sources\n\n", t, func() {

        cfg := struct {
            Host        string      `input:"env,flag,term" usage:"Server host name"`
            Port        int         `input:"env,flag,term" usage:"Server port number" min:"1" max:"65535"`
            Enabled     bool        `input:"env,flag,term" usage:"Server enabled"`
            Admin       string      `input:"flag,term"     usage:"Administrator account"`
            Password    string      `input:"flag,term,noecho" usage:"Administrator password"`
        }{
            Host:       "localhost.localdomain",
            Port:       5432,
            Enabled:    true,
            Admin:      "pgadmin",
        }

        ui := NewUserInput(&cfg)
        ui.SetHelpText("Show help text with command line options.\n" +
            "The following option list has been generated from the configured struct tags.")

        err := ui.BindInput()
        So(err, ShouldBeNil)
        So(ui.HelpRequested(), ShouldBeTrue)

    })

    Convey("Show help, terminal input only\n\n", t, func() {

        cfg := struct {
            Admin       string      `input:"term"        usage:"Administrator account"`
            Password    string      `input:"term,noecho" usage:"Administrator password"`
        }{}

        ui := NewUserInput(&cfg)
        err := ui.BindInput()
        So(err, ShouldBeNil)
        So(ui.HelpRequested(), ShouldBeTrue)

    })

}



/***/