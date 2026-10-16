package userinput_tests

import (
  u "userinput"

    "os"
    "fmt"
    "os/exec"
    "errors"
    //"strings"


    //"math"
    //"flag"
    //"bytes"
    //r "reflect"
    "testing"
  . "github.com/smartystreets/goconvey/convey"
)

const  DISPLAY_TEST_OUTPUT = true

func exec_TestHelper (run string) (string, int) {
    cmd := exec.Command( os.Args[0], "-test.run=^"+run+"$" )
    cmd.Env = append( os.Environ(), "USERINPUT_TEST_HELPER=1" )
    out, err := cmd.CombinedOutput()
    output := string(out)
    var exitErr *exec.ExitError
    if !errors.As(err, &exitErr) {
        return output, 0
    }
    return output, exitErr.ExitCode()
}

func display_output (s string) {
    if DISPLAY_TEST_OUTPUT {
        fmt.Println("\n*** ***  Output  *** ***\n" +
        s + "\n*** *** *** *** *** ***\n")
    }
}




// // // // // // // // // // // // // //
//
func Test_FlagEnv030 (t *testing.T) {

    Convey("Flag and Env Errors 030: value out of range\n", t, func() {
        output, ExitCode := exec_TestHelper("TestHelper_FlagEnv030")
        display_output(output)

        So(ExitCode, ShouldEqual, 2)
        So(output, ShouldContainSubstring, "value out of range, for environment variable NUMBER_A, expecting 10 to 20")
        So(output, ShouldContainSubstring, "value out of range (9), for flag -number-a, expecting 10 to 20")
    })

    Convey("Flag and Env Errors 031: invalid integer\n", t, func() {
        output, ExitCode := exec_TestHelper("TestHelper_FlagEnv031")
        display_output(output)

        So(ExitCode, ShouldEqual, 2)
        So(output, ShouldContainSubstring, `invalid value for environment variable NUMBER_B`)
        So(output, ShouldContainSubstring, `invalid value "nine" for flag -number-b: invalid integer`)
    })

    Convey("Flag and Env Errors 032: invalid boolean\n", t, func() {
        output, ExitCode := exec_TestHelper("TestHelper_FlagEnv032")
        display_output(output)

        So(ExitCode, ShouldEqual, 2)

        So(output, ShouldContainSubstring, `invalid boolean value for environment variable ENABLED, expecting true or false`)
        So(output, ShouldContainSubstring, `invalid boolean value "yerp" for -enabled: expecting true or false`)
    })

    Convey("Flag and Env Errors 033: invalid string\n", t, func() {
        output, ExitCode := exec_TestHelper("TestHelper_FlagEnv033")
        display_output(output)

        So(ExitCode, ShouldEqual, 2)

        So(output, ShouldContainSubstring, `invalid value for environment variable BNAME`)
        So(output, ShouldContainSubstring, `invalid string "Fail this!", for flag -bname`)
    })

}

func TestHelper_FlagEnv030 (t *testing.T) {
    if os.Getenv("USERINPUT_TEST_HELPER") != "1" {
        return
    }
    os.Args = []string{
        "mycommand",
        "-number-a=9",
    }
    os.Setenv("NUMBER_A",       "9")

    var cfg struct {
        Number_A        int     `input:"env,flag"  min:"10"  max:"20"`
    }
    ui := u.NewUserInput(&cfg)
    ui.BindInput()

    t.Fail() // Should not get to this line
}

func TestHelper_FlagEnv031 (t *testing.T) {
    if os.Getenv("USERINPUT_TEST_HELPER") != "1" {
        return
    }
    os.Args = []string{
        "mycommand",
        "-number-b=nine",
    }
    os.Setenv("NUMBER_B",       "nine")

    var cfg struct {
        Number_B        int     `input:"env,flag"`
    }
    ui := u.NewUserInput(&cfg)
    ui.BindInput()
    t.Fail() // Should not get to this line
}

func TestHelper_FlagEnv032 (t *testing.T) {
    if os.Getenv("USERINPUT_TEST_HELPER") != "1" {
        return
    }
    os.Args = []string{
        "mycommand",
        "-enabled=yerp",
    }
    os.Setenv("ENABLED",        "yerp")

    var cfg struct {
        Enabled         bool    `input:"env,flag"`
    }
    ui := u.NewUserInput(&cfg)
    ui.BindInput()
    t.Fail() // Should not get to this line
}

func TestHelper_FlagEnv033 (t *testing.T) {
    if os.Getenv("USERINPUT_TEST_HELPER") != "1" {
        return
    }
    os.Args = []string{
        "mycommand",
        "-bname=Fail this!",
    }
    os.Setenv("BNAME",          "Fail this!")

    var cfg struct {
        B               string  `input:"env,flag"  name:"bname"  regexp:"^[a-zA-z]+$"`
    }
    ui := u.NewUserInput(&cfg)
    ui.BindInput()
    t.Fail() // Should not get to this line
}

/**/
