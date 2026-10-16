package userinput

// Environment-variable binding.

import (
    "fmt"
    "os"
  r "reflect"
    "strconv"
)

// bindEnv applies environment variables to configured fields. Missing variables
// are ignored. Invalid values are returned as a slice of errors.
func (ui *UserInput) bindEnv() []error {

    var errs []error

    for fieldName, cf := range ui.vogons {

        // skip if no env name
        if cf.envName == "" {
            continue
        }

        // Skip variables that are not present in the process environment.
        input, ok := os.LookupEnv(cf.envName)
        if !ok {
            continue
        }

        switch cf.fieldValue.Kind() {
        case r.Bool:
            // Boolean environment variables accept case-insensitive 1, 0, t, f, true, false
            value, err := strconv.ParseBool(input)
            if err != nil {
                errs = append(errs, fmt.Errorf(`invalid boolean value for environment variable %s, expecting true or false`, cf.envName))
                continue
            }
            // Store the validated boolean in the staging field.
            cf.stagedValue.SetBool(value)

            // A variable is considered supplied only after its value has
            // been successfully validated and stored.
            ui.supplied[fieldName] = true

        case r.Int:
            // Validate integer
            n, err := strconv.Atoi(input)
            if err != nil {
                errs = append(errs, fmt.Errorf(`invalid value for environment variable %s`, cf.envName))
                continue
            }

    		if n < cf.min || n > cf.max {
                errs = append(errs, fmt.Errorf("value out of range, for environment variable %s, expecting %d to %d", cf.envName, cf.min, cf.max))
                continue
            }

            cf.stagedValue.SetInt(int64(n))
            ui.supplied[fieldName] = true

        case r.String:
            // validate string
            if cf.re != nil && input != "" {
                if !cf.re.MatchString(input) {
                    errs = append(errs, fmt.Errorf(`invalid value for environment variable %s`, cf.envName))
                    continue
                }
            }

            cf.stagedValue.SetString(input)
            ui.supplied[fieldName] = true

        }
    }

    return errs

}
