package userinput

import (
	"bufio"
	"errors"
	"os"

	"fmt"
	"golang.org/x/term"
	r "reflect"
	"strconv"
	"strings"
)

// bindTerminal prompts for fields configured for terminal input that have not
// already been recorded in ui.supplied. A blank response keeps the current value.
func (ui *UserInput) bindTerminal() error {

	for fieldName, cf := range ui.vogons {

		// skip if no terminal input
		if !cf.termInput {
			continue
		}

		// skip if already supplied
		if _, ok := ui.supplied[fieldName]; ok {
			continue
		}

		switch cf.fieldValue.Kind() {
		case r.Bool:
			err := ui.boolTermInput(cf)
			if err != nil {
				return err
			}

		case r.Int:
			err := ui.intTermInput(cf)
			if err != nil {
				return err
			}

		case r.String:
			err := ui.stringTermInput(cf)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

// stringTermInput reads and validates one string field from the terminal.
func (ui *UserInput) stringTermInput(cf *ConfiguredField) error {

   	prompt := cf.prompt

	var mode InputMode = NormalEcho
	if cf.suppressEcho {
		mode = SuppressEcho
	} else {
    	if !cf.fieldValue.IsZero() {
    		prompt = fmt.Sprintf("%s [%s]", prompt, cf.fieldValue.String())
    	}
    }

	for {

		input, err := ttyReadLine(prompt, mode)
		if err != nil {
			return err
		}

		if input == "" {
			break
		}

		if cf.re != nil && !cf.re.MatchString(input) {
			ttyPrintf("Invalid input\n")
			continue
		}

		ui.supplied[cf.fieldName] = true
		cf.stagedValue.SetString(input)
		break
	}

	return nil
}

// intTermInput reads an integer until a blank or valid in-range value is entered.
func (ui *UserInput) intTermInput (cf *ConfiguredField) error {

   	prompt := cf.prompt
	var mode InputMode = NormalEcho

	if cf.suppressEcho {
		mode = SuppressEcho
	} else {
    	if !cf.fieldValue.IsZero() {
    		prompt = fmt.Sprintf("%s [%d]", prompt, cf.fieldValue.Int())
    	}
	}

	for {

		input, err := ttyReadLine(prompt, mode)
		if err != nil {
			return err
		}

		if input == "" {
			break
		}

		n, err := strconv.ParseInt(input, 10, strconv.IntSize)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				ttyPrintf("Integer overflow")
			}
			ttyPrintf("Invalid integer")
			continue
		}

		value := int(n)

		if value < cf.min || value > cf.max {
			ttyPrintf("Value out of range (%d), expecting %d - %d\n", value, cf.min, cf.max)
			continue
		}

		ui.supplied[cf.fieldName] = true
		cf.stagedValue.SetInt(int64(value))

		break
	}

	return nil
}

// boolTermInput accepts y or n, with a blank response preserving the current value.
func (ui *UserInput) boolTermInput(cf *ConfiguredField) error {

   	prompt := cf.prompt
	dfault := "y/N"

	if cf.fieldValue.Bool() == true {
		dfault = "Y/n"
	}
	prompt = fmt.Sprintf("%s [%s]", prompt, dfault)

	for {

		input, err := ttyReadLine(prompt, NormalEcho)
		if err != nil {
			return err
		}

		if input == "" {
			break
		}

		input = strings.ToLower(input)

		if !validateString(input, `^[yn]?$`) {
			ttyPrintf("Invalid input\n")
			continue
		}

		ui.supplied[cf.fieldName] = true
		if input == "y" {
			cf.stagedValue.SetBool(true)
		} else {
			cf.stagedValue.SetBool(false)
		}

		break
	}

	return nil
}

// ttyReadLine is replaceable so terminal input can be exercised without an
// interactive terminal during tests. The default implementation reads /dev/tty.
var ttyReadLine = func(prompt string, mode InputMode) (line string, err error) {

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer tty.Close()

	fmt.Fprintf(tty, "%s: ", prompt)

	if mode == SuppressEcho {
		b, err := term.ReadPassword(int(tty.Fd()))
		if err != nil {
			return "", err
		}
		line = string(b)
		fmt.Fprintf(tty, "\n")
	} else {
		line, err = bufio.NewReader(tty).ReadString('\n')
		if err != nil {
			return "", err
		}
	}

	line = strings.TrimRight(line, "\r\n")

	return line, nil
}

// TermReadPassword reads a password from /dev/tty without echoing the value.
// It writes the supplied prompt to the terminal and returns the entered text.
func TermReadPassword(prompt string) (string, error) {

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer tty.Close()

	fmt.Fprint(tty, prompt)

	b, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprint(tty, "\n")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// TermReadInput reads text from /dev/tty.
// It writes the supplied prompt to the terminal and returns the entered text.
func TermReadInput(prompt string) (string, error) {

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer tty.Close()

	fmt.Fprint(tty, prompt)

	line, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimRight(line, "\r\n")
	return line, nil
}

// ttyPrintf writes terminal validation feedback directly to /dev/tty.
func ttyPrintf(format string, a ...any) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer tty.Close()

	fmt.Fprintf(tty, format, a...)
}
