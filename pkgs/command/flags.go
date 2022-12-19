package command

import (
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnolang/gno/pkgs/errors"
)

var reFlagName = regexp.MustCompile(`^--[a-z0-9.\-]+(#[a-z0-9.\-]+)?$`)

// applies all flags to ptr to options.
// --flag is short for --flag true for boolean flags.
// consecutive flags can be used to populate arrays or slices.
// alternatively, a comma on a single flag can be used.
func applyFlags(ptr interface{}, flags map[string]interface{}) error {
	prv := reflect.ValueOf(ptr)
	if prv.Type().Kind() != reflect.Ptr {
		panic("expected pointer kind to option")
	}
	rv := prv.Elem()
	return applyFlagsReflect(rv, flags)
}

// apply all flags or return error.
func applyFlagsReflect(rv reflect.Value, flags map[string]interface{}) error {
	for fname, fvalue := range flags {
		match, err := applyFlagReflect(rv, fname, fvalue)
		if err != nil {
			return err
		} else if match {
			continue
		} else {
			return errors.New("no field found with flag name %s", fname)
		}
	}
	// all matched, return no error.
	return nil
}

// apply flag with name fname to struct or field.
func applyFlagReflect(rv reflect.Value, fname string, fvalue interface{}) (bool, error) {
	// scan all fields to find match.
	// NOTE inefficient.
	// TODO cache/index fields by name.
	rt := rv.Type()
	num := rv.NumField()
	for i := 0; i < num; i++ {
		rtf := rt.Field(i)
		ffn := rtf.Tag.Get("flag")
		if rtf.Anonymous {
			// try to match, otherwise continue with other fields.
			frv := rv.Field(i)
			match, err := applyFlagReflect(frv, fname, fvalue)
			if err != nil {
				return false, err
			} else if match {
				// found match, done!
				return true, nil
			} else {
				// continue
			}
		} else if ffn == "" {
			// ignore fields with no flags field.
			// NOTE: instead of returning an error here,
			// check all structs for consistency beforehand instead.
			// Otherwise it's "offensive" programming.
			fmt.Fprintf(os.Stderr, "WARN: non-anonymous option field found (%s) with no flag name; in the future this will panic at start of program\n", rtf.Name)
		} else if ffn == fname {
			frv := rv.Field(i)
			return true, applyFlagToFieldReflect(frv, fvalue)
		}
	}
	return false, nil
}

// apply flag value to a matched field.
func applyFlagToFieldReflect(frv reflect.Value, fvalue interface{}) error {
	switch cfvalue := fvalue.(type) {
	case map[string]interface{}:
		if frv.Type().Kind() != reflect.Struct {
			return errors.New(
				"expected struct kind but got %v",
				frv.Type())
		}
		return applyFlagsReflect(frv, cfvalue)
	case string:
		return applyFlagToFieldReflectString(frv, cfvalue)
	case []string:
		return applyFlagToFieldReflectStringSlice(frv, cfvalue)
	default:
		panic("should not happen")
	}
}

// apply flag value string to a matched field.
func applyFlagToFieldReflectString(frv reflect.Value, fvalue string) error {
	frt := frv.Type()
	switch frt.Kind() {
	case reflect.Ptr:
		if frv.IsNil() {
			frv.Set(reflect.New(frt.Elem()))
		}
		err := applyFlagToFieldReflectString(frv.Elem(), fvalue)
		return err
	case reflect.Array:
		ert := frt.Elem()
		if ert.Kind() == reflect.Uint8 {
			bz, err := hex.DecodeString(fvalue)
			if err != nil {
				// if not hex, try to use the fvalue directly.
				bz = []byte(fvalue)
				// return errors.Wrap(err, "invalid hex")
			}
			frv.SetBytes(bz)
			return nil
		} else {
			parts := strings.Split(fvalue, ",")
			for i, part := range parts {
				erv := frv.Index(i)
				err := applyFlagToFieldReflectString(erv, part)
				if err != nil {
					return errors.Wrap(err, "error parsing item")
				}
			}
			return nil
		}
	case reflect.Slice:
		ert := frt.Elem()
		if ert.Kind() == reflect.Uint8 {
			bz, err := hex.DecodeString(fvalue)
			if err != nil {
				// if not hex, try to use the fvalue directly.
				bz = []byte(fvalue)
				// return errors.Wrap(err, "invalid hex")
			}
			frv.SetBytes(bz)
			return nil
		} else {
			parts := strings.Split(fvalue, ",")
			srv := reflect.MakeSlice(frt, len(parts), len(parts))
			frv.Set(srv)
			for i, part := range parts {
				erv := frv.Index(i)
				err := applyFlagToFieldReflectString(erv, part)
				if err != nil {
					return errors.Wrap(err, "error parsing item")
				}
			}
			return nil
		}
	case reflect.Int:
		fnum, err := strconv.ParseInt(fvalue, 0, 0)
		if err != nil {
			return errors.Wrap(err, "invalid int")
		}
		frv.SetInt(fnum)
		return nil
	case reflect.Int8:
		fnum, err := strconv.ParseInt(fvalue, 0, 8)
		if err != nil {
			return errors.Wrap(err, "invalid int8")
		}
		frv.SetInt(fnum)
		return nil
	case reflect.Int16:
		fnum, err := strconv.ParseInt(fvalue, 0, 16)
		if err != nil {
			return errors.Wrap(err, "invalid int16")
		}
		frv.SetInt(fnum)
		return nil
	case reflect.Int32:
		fnum, err := strconv.ParseInt(fvalue, 0, 32)
		if err != nil {
			return errors.Wrap(err, "invalid int32")
		}
		frv.SetInt(fnum)
		return nil
	case reflect.Int64:
		fnum, err := strconv.ParseInt(fvalue, 0, 64)
		if err != nil {
			return errors.Wrap(err, "invalid int64")
		}
		frv.SetInt(fnum)
		return nil
	case reflect.Uint:
		fnum, err := strconv.ParseUint(fvalue, 0, 0)
		if err != nil {
			return errors.Wrap(err, "invalid uint")
		}
		frv.SetUint(fnum)
		return nil
	case reflect.Uint8:
		fnum, err := strconv.ParseUint(fvalue, 0, 8)
		if err != nil {
			return errors.Wrap(err, "invalid uint8")
		}
		frv.SetUint(fnum)
		return nil
	case reflect.Uint16:
		fnum, err := strconv.ParseUint(fvalue, 0, 16)
		if err != nil {
			return errors.Wrap(err, "invalid uint16")
		}
		frv.SetUint(fnum)
		return nil
	case reflect.Uint32:
		fnum, err := strconv.ParseUint(fvalue, 0, 32)
		if err != nil {
			return errors.Wrap(err, "invalid uint32")
		}
		frv.SetUint(fnum)
		return nil
	case reflect.Uint64:
		fnum, err := strconv.ParseUint(fvalue, 0, 64)
		if err != nil {
			return errors.Wrap(err, "invalid uint64")
		}
		frv.SetUint(fnum)
		return nil
	case reflect.String:
		// XXX is there something wrong with os.Args? why does it strip '/", and then not unescape \n and \t while unescaping \\?
		fvalue = strings.ReplaceAll(fvalue, `\n`, "\n")
		fvalue = strings.ReplaceAll(fvalue, `\t`, "\t")
		frv.SetString(fvalue)
		return nil
	case reflect.Bool:
		switch fvalue {
		case "true", "True", "yes", "Yes", "y", "Y":
			frv.SetBool(true)
			return nil
		case "false", "False", "no", "No", "n", "N":
			frv.SetBool(false)
			return nil
		default:
			return errors.New("unexpected bool value: " + fvalue)
		}
	case reflect.Struct:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"flag value cannot be applied to field of type %s",
			frt.String()))

	}
}

func applyFlagToFieldReflectStringSlice(frv reflect.Value, fvalues []string) error {
	frt := frv.Type()
	switch frt.Kind() {
	case reflect.Array:
		for i, part := range fvalues {
			erv := frv.Index(i)
			err := applyFlagToFieldReflectString(erv, part)
			if err != nil {
				return errors.Wrap(err, "error parsing item")
			}
		}
		return nil
	case reflect.Slice:
		srv := reflect.MakeSlice(frt, len(fvalues), len(fvalues))
		frv.Set(srv)
		for i, part := range fvalues {
			erv := frv.Index(i)
			err := applyFlagToFieldReflectString(erv, part)
			if err != nil {
				return errors.Wrap(err, "error parsing item")
			}
		}
		return nil
	default:
		panic(fmt.Sprintf(
			"flag values cannot be applied to field of type %s",
			frt.String()))

	}
}

// all flags follow non-flag args.
func ParseArgs(oargs []string) (args []string, flags map[string]interface{}) {
	for i, arg := range oargs {
		if strings.HasPrefix(arg, "-") {
			args = oargs[:i]
			flags = parseFlags(oargs[i:])
			return
		}
	}
	args = oargs
	flags = nil
	return
}

func parseFlags(fargs []string) map[string]interface{} {
	if len(fargs) == 0 {
		return nil
	}
	m := make(map[string]interface{}, len(fargs))
	var fnamePrev string // for keeping track of repeated flags.
	var fname string
	for _, farg := range fargs {
		if strings.HasPrefix(farg, "--") {
			if fname != "" {
				// is --flag shortform (like --flag true).
				// this cannot happen with repeated flags.
				if fnamePrev == fname {
					panic(fmt.Sprintf(
						"repeated flags cannot include implicit true boolean"))
				}
				// this cannot happen with file flags.
				if strings.HasSuffix(fname, "#file") {
					panic(fmt.Sprintf(
						"file name not provided for " + fname))
				}
				// set y for yes.
				setFlag(m, fname, "y", false)
			}
			fname = parseFlagName(farg)
		} else {
			if fname == "" {
				panic(fmt.Sprintf(
					"dangling flag value in args: %s",
					farg))
			}
			// if a --flag#file <file_location> flag, read contents.
			if strings.HasSuffix(fname, "#file") {
				ffile := farg
				fargbz, err := os.ReadFile(ffile)
				if err != nil {
					panic(fmt.Sprintf(
						"error reading file: %v", err))
				}
				// update fname and farg.
				fname = fname[:len(fname)-len("#file")]
				farg = string(fargbz)
			}
			repeat := fname == fnamePrev
			setFlag(m, fname, farg, repeat)
			fnamePrev = fname // remember
			fname = ""        // reset
		}
	}
	if fname != "" {
		// trailing --fname
		repeat := fname == fnamePrev
		setFlag(m, fname, "y", repeat) // y for yes
	}
	return m
}

// Set the flag value of a key identified by fname to m.
// If fname contains a dot, m will contain a nested map.
// If repeat is true, fvalue will be appended to a slice of existing arg(s).
// Otherwise, panics when encountering a pre-existing flag.
func setFlag(m map[string]interface{}, fname string, fvalue string, repeat bool) {
	parts := strings.Split(fname, ".")
	setFlagWithParts(m, fname, parts, fvalue, repeat)
}

// fname: the original flag name.
func setFlagWithParts(m map[string]interface{}, fname string, fparts []string, fvalue string, repeat bool) {
	if len(fparts) > 1 {
		first := fparts[0]
		if m2i, ok := m[first]; ok {
			m2 := m2i.(map[string]interface{})
			setFlagWithParts(m2, fname, fparts[1:], fvalue, repeat)
		} else {
			m2 := make(map[string]interface{})
			setFlagWithParts(m2, fname, fparts[1:], fvalue, repeat)
			m[first] = m2
		}
	} else {
		name := fparts[0]
		if !repeat {
			if _, exists := m[name]; exists {
				panic(fmt.Sprintf(
					"flag already set: %s (and repeated flags must be consecutive)", fname))
			}
			m[name] = fvalue
		} else {
			fvaluePrev, exists := m[name]
			if !exists {
				panic("should not happen")
			}
			switch fvaluePrev.(type) {
			case string:
				m[name] = []string{fvaluePrev.(string), fvalue}
			case []string:
				m[name] = append(fvaluePrev.([]string), fvalue)
			default:
				panic("should not happen")
			}
		}
	}
}

func parseFlagName(farg string) string {
	match := reFlagName.MatchString(farg)
	if !match {
		panic(fmt.Sprintf(
			"invalid flag name: %s",
			farg))
	}
	return farg[2:]
}
