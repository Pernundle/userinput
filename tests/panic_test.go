package userinput_tests

import (
  u "userinput"

    //"os"
    //"fmt"
    //"flag"
    //"bytes"
  //r "reflect"
    "testing"
  . "github.com/smartystreets/goconvey/convey"
)


//  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //
func Test_Panic_Config (t *testing.T) {

    cfg := make(map[string]string)

    Convey("Convey Panic - UserInput not a pointer:\n", t, func() {

        panics := func(){
            _ = u.NewUserInput(cfg)
        }

        So(panics, ShouldPanicWith, "cfg must be a non-nil pointer to a struct")

    })

    Convey("Convey Panic - UserInput not a pointer to a struct:\n", t, func() {

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "cfg must be a pointer to a struct")

    })

}


//  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //  //
func Test_Config_Panic (t *testing.T) {

    Convey("Convey Panic - Unsupported type:\n", t, func() {

        var cfg struct {
            BadField     uint  `input:"flag"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "Kind not supported: uint")

    })

    Convey("Convey Panic - Duplicate Name:\n", t, func() {

        var cfg struct {
            MyField1     int  `input:"flag"  name:"myField2"`
            MyField2     int  `input:"flag"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "Duplicate name: myfield2 at field MyField2")

    })

    Convey("Convey Panic - noecho bool:\n", t, func() {

        type cfgBool struct {
            MyField1     bool  `input:"term,noecho"`
        }

        var cfg cfgBool

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "cannot use 'noecho' with type bool: MyField1")

    })

    Convey("Convey Panic - Undefined Input Source:\n", t, func() {

        var cfg struct {
            MyField1     string  `input:"term,flagg"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "undefined input source: flagg in field MyField1")

    })

    Convey("Convey Panic - Empty regexp tag:\n", t, func() {

        var cfg struct {
            MyField1     string  `input:"term" regexp:""`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "tag regexp cannot be empty: MyField1")

    })

    Convey("Convey Panic - Regexp tag on non-string field:\n", t, func() {

        var cfg struct {
            MyField1     int  `input:"term" regexp:"[0-9]*"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "tag regexp can only be used with type string: MyField1")

    })

    Convey("Convey Panic - Malformed Regexp:\n", t, func() {

        var cfg struct {
            MyField1     string  `input:"term" regexp:"[a-z"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanic)

    })

    Convey("Convey Panic - Min tag on non integer field:\n", t, func() {

        var cfg struct {
            MyField1     string  `input:"term" min:"0"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "tag min can only be used with type int: MyField1")

    })

    Convey("Convey Panic - Max tag on non integer field:\n", t, func() {

        var cfg struct {
            MyField1     string  `input:"term" max:"0"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "tag max can only be used with type int: MyField1")

    })

    Convey("Convey Panic - Invalid min tag:\n", t, func() {

        var cfg struct {
            MyField1     int     `input:"term" min:"a"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, `Field: MyField1, tag: min, invalid syntax`)

    })

    Convey("Convey Panic - Invalid max tag:\n", t, func() {

        var cfg struct {
            MyField1     int     `input:"term" max:"a"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, `Field: MyField1, tag: max, invalid syntax`)

    })

    Convey("Convey Panic - Min tag greater than max tag:\n", t, func() {

        var cfg struct {
            MyField1     int     `input:"term" min:"1" max:"0"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "Field: MyField1, min tag is greater than max tag")

    })

    Convey("Convey Panic - No exported field:\n", t, func() {

        var cfg struct {
            myField1     int     `input:"term"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "No exported field was found in cfg")

    })

    Convey("Convey Panic - No Input Tag:\n", t, func() {

        var cfg struct {
            MyField1     int     `name:"myfield1"`
        }

        panics := func(){
            _ = u.NewUserInput(&cfg)
        }

        So(panics, ShouldPanicWith, "No field was found with an `input` tag")

    })

}

