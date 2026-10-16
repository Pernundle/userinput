package userinput

import (
	//"os"
	"fmt"
	"math"
    "strings"

	r "reflect"
	"regexp"
	"strconv"
)


// getPrompt returns the usage tag, or the field name with "Enter" prepended when usage is omitted.
func getPrompt(f r.StructField) string {
	if s, ok := f.Tag.Lookup("usage"); ok {
		return s
	}
	return "Enter " + f.Name
}

// getName returns the validated name tag, or the field name when no alias is set.
func getName(f r.StructField) string {
	alias, ok := f.Tag.Lookup("name")
	if ok {
		if !compiledNameRegexp.MatchString(alias) {
			panic(fmt.Sprintf("invalid name tag (%s) in field: %s", alias, f.Name))
		}
	} else {
		alias = f.Name
	}
	return strings.ToLower(alias)
}

// getUsage returns the usage tag, or the field name when usage is omitted.
func getUsage(f r.StructField) string {
	if s, ok := f.Tag.Lookup("usage"); ok {
		return s
    }
	return f.Name
}

// findStringSubmatch returns the first captured submatch from regx.
func findStringSubmatch(s, regx string) (string, bool) {
	match := regexp.MustCompile(regx).FindStringSubmatch(s)
	if len(match) > 1 {
		return match[1], true
	}
	return "", false
}

// validateString reports whether s matches regx.
func validateString (s, regx string) bool {
	return regexp.MustCompile(regx).MatchString(s)
}

// supportedKind reports whether v can be bound by this package.
func supportedKind(v r.Value) bool {
	switch v.Kind() {
	case r.Bool, r.Int, r.String:
		return true
	}
	return false
}

// getMinMax reads integer range tags and panics when the tags are invalid.
func getMinMax (cf *ConfiguredField) {
	cf.min = math.MinInt
	cf.max = math.MaxInt

    field := cf.fieldInfo
    var err error

	if tagMin, ok := field.Tag.Lookup("min"); ok {
		cf.min, err = strconv.Atoi(tagMin)
		if err != nil {
			panic(fmt.Sprintf("Field: %s, tag: min, invalid syntax", field.Name))
		}
        cf.minSet = true
	}
	if tagMax, ok := field.Tag.Lookup("max"); ok {
		cf.max, err = strconv.Atoi(tagMax)
		if err != nil {
			panic(fmt.Sprintf("Field: %s, tag: max, invalid syntax", field.Name))
		}
        cf.maxSet = true
	}
	if cf.min > cf.max {
		panic(fmt.Sprintf("Field: %s, min tag is greater than max tag", field.Name))
	}
}

/* *** *** *** *** *** *** *** *** *** *** *** *** *** *** *** *** *** *** */

// intValue adapts a reflected integer field to flag.Value.
//
// The reflected field remains connected to the original cfg struct.
// Consequently, Set changes the original struct field rather than
// storing a separate copy of its value.
type intValue struct {
	field r.Value
}

// String returns the current value in its command-line representation.
// flag.Var uses this when recording and displaying the flag's default value.
//
// The nil and invalid checks are required because the flag package may call
// String on a zero-valued Value while formatting usage information.
func (i *intValue) String () string {
	if i == nil || !i.field.IsValid() {
		return "0"
	}
	return strconv.FormatInt(i.field.Int(), 10)
}

// Set converts the command-line string to an integer and stores it directly
// in the reflected struct field. SetInt works for both int and named types
// whose underlying kind is reflect.Int.
func (v *intValue) Set (s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
	    return fmt.Errorf("invalid integer")
	}
	v.field.SetInt(int64(n))
	return nil
}

// // // // // // // // // //
// stringValue adapts a reflected string field to flag.Value.
type stringValue struct {
	field r.Value
}

// String returns the current value in its command-line representation.
// flag.Var uses this when recording and displaying the flag's default value.
func (s *stringValue) String () string {
	if s == nil || !s.field.IsValid() {
		return ""
	}
	return s.field.String()
}

// Set stores the command-line value directly in the reflected struct field.
func (v *stringValue) Set (s string) error {
	v.field.SetString(s)
	return nil
}

// // // // // // // // // //
// boolValue adapts a reflected boolean field to flag.Value.
type boolValue struct {
	field r.Value
}

// String returns the current value in its command-line representation.
// flag.Var uses this when recording and displaying the flag's default value.
func (b *boolValue) String () string {
	if b == nil || !b.field.IsValid() {
		return "false"
	}
	return strconv.FormatBool(b.field.Bool())
}

// Set parses the command-line value as a boolean and stores it directly
// in the reflected struct field.
func (b *boolValue) Set (s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
	    return fmt.Errorf("expecting true or false")
	}
	b.field.SetBool(v)
	return nil
}

func (v *boolValue) IsBoolFlag() bool {
    return true
}

// Documentation written by Artificial Intelligence and edited by Alleged Inelligence.
