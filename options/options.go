// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

/*
Package options provides a command line option parser.

This package is meant as an alternative to the core flag package. It
is more powerful without attempting to support every possible feature
some parsing library ever introduced. It is arguably easier to use.

Usage:

Create an OptionSpec that documents your program's allowed flags. This
begins with a free-text synopsis of your command line interface, then
a line containing only two dashes, then a series of option specifications:

  import "github.com/gaal/go-options/options"
  s := options.NewOptions(`
  cat - concatenate files to standard input
  Usage: cat [OPTIONS] file...
  This version of cat supports character set conversion.
  Fancifully, you can say "-r 3" and have everything told you three times.
  --
  n,numerate,number     number input lines
  e,escape              escape nonprintable characters
  i,input-encoding=     charset input is encoded in [utf-8]
  o,output-encoding=    charset output is encoded in [utf-8]
  r,repeat=             repeat every line some number of times [1]
  v,verbose             be verbose
  `)

Then parse the command line:

  opt := s.Parse(os.Args[1:])

This return value is of Options type and has the parsed command line (for
an alternate way of parsing the command line, see the "Callback interface"
descibed below).

Options may have any number of aliases; the last one is the "canonical"
name and the one your program must use when reading values.

  opt.Get("input-encoding")  // Returns "utf-8", or whatever user set.
  opt.Get("i")               // Error! No option with that canonical name.
  opt.Get("number")          // Returns "" if the user didn't specify it.

Get returns a string. Several very simple conversions are provided but you
are encouraged to write your own if you need more.

  opt.GetBool("escape")      // false (by default)
  opt.GetBool("number")      // false (by default)
  opt.GetInt("repeat")       // 1 (by default)

Options either take a required argument or take no argument. Non-argument
options have useful values exposed as bool and ints.

  // cat -v -v -v, or alternatively,
  // cat -vvv
  opt.GetBool("verbose")     // true
  opt.GetInt("verbose")      // 3

The user can say either "--foo=bar" or "--foo bar". Short options may be
clustered; "-abc foo" means the same as "-a -b -c=foo".

Parsing stops if "--" is given on the command line.

The "Extra" field of the returned Options contains all non-option command line
input. In the case of a cat command, this would be the filenames to concat.

By default, options permits such extra values. Setting UnknownValuesFatal
causes it to panic when it enconters them instead.

The "Flags" field of the returned Options contains the series of flags as given
on the command line, including repeated ones (which are suppressed in opt --
it only contains the last value). This allows you to do your own handling
of repeated options easily.

By default, options does not permit unknown flags. Setting
UnknownOptionsFatal to false causes them to be recorded in "flags" instead.
Note that since they have no canonical name, they cannot be accessed via
opt. Also note that since options does not know about the meaning of these
flags, it has to guess whether they consume the next argument or not. This
is currently done naively by peeking at the first character of the next
argument.

Callback interface:

If you prefer a more type-safe, static interface to your options, you can
use options to get that too. Instead of looking at opt and friends, set
OptionSpec.ParseCallback:

  var (foo string; bar int; baz float64; lst []string; verbose bool)

  func myParseCallback(spec *OptionSpec, option string, argument *string) {
    if argument != nil {
      switch option {
      case "string-option":  foo = *argument
      case "int-option":     fmt.Sscanf(*argument, "%d", &bar)
      case "decimal-option": fmt.Sscanf(*argument, "%f", &baz)
      case "list-option":    lst = append(lst, *argument)
      default: spec.PrintUsageAndExit("Unknown option: " + option)
      }
    } else {
      switch option {
      case "verbose": verbose = true
      case "help":    spec.PrintUsageAndExit("")  // No error
      default:        spec.PrintUsageAndExit("Unknown option: " + option)
    }
  }

  spec.ParseCallback = myParseCallback
  opt := spec.Parse(os.Args[1:])
  // Note that the opt.Get won't work when using a custom parse callback.

*/
package options

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const EX_USAGE = 64 // Exit status for incorrect command lines.

// Options represents the known formal options provided on the command line.
type Options struct {
	opts     map[string]string
	known    map[string]bool
	Flags    [][]string // Original flags presented on the command line
	Extra    []string   // Non-option command line arguments left on the command line
	Leftover []string   // Untouched arguments (after "--")
}

// Get returns the value of an option, which must be known to this parse.
// Options that take an argument return the argument. Options with no argument
// return values hinting whether they were specified or not; GetInt or GetBool
// may be more suited for looking them up.
func (o *Options) Get(flag string) string {
	val, ok := o.opts[flag]
	if !ok && !o.known[flag] {
		panic(fmt.Sprintf("[Programmer error] Unknown option: %s\ndump: %+v", flag, *o))
	}
	return val
}

// GetInt returns the value of an option as an integer. The empty string is
// treated as zero, but otherwise the option must parse or a panic occurs.
func (o *Options) GetInt(flag string) int {
	val := o.Get(flag)
	if val == "" {
		return 0
	}
	var num int
	if n, _ := fmt.Sscan(val, &num); n != 1 {
		panic("Bad integer value for option: " + flag + ": " + o.Get(flag))
	}
	return num
}

// GetBool returns the value of an option as a bool. All values are treated
// as true except for the following which yield false:
//   "" (empty), "0", "false", "off", "nil", "null", "no"
func (o *Options) GetBool(flag string) bool {
	val := o.Get(flag)
	if val == "" || val == "0" || val == "false" ||
		val == "off" || val == "nil" || val == "null" || val == "no" {
		return false
	}
	return true
}

// Have returns false when an option has no default value and was not given
// on the command line, or true otherwise.
func (o *Options) Have(flag string) bool {
	if !o.known[flag] {
		panic(fmt.Sprintf("[Programmer error] Unknown option: %s\ndump: %+v", flag, *o))
	}
	_, ok := o.opts[flag]
	return ok
}

// GetAll is a convenience function which scans the "flags" return value of
// OptionSpec.Parse, and gathers all the values of a given option. This must
// be a required-argument option.
func GetAll(flag string, flags [][]string) []string {
	out := make([]string, 0)
	for _, val := range flags {
		if val[0] == flag {
			if len(val) != 2 {
				panic("[Programmer error] Option does not appear to take arguments: " + flag)
			}
			out = append(out, val[1])
		}
	}
	return out
}

// BUG(gaal): Option groups are not yet supported.

// BUG(gaal): The usage string is not yet formatted prettily. It should consider tty width, etc.

// OptionSpec represents the specification of a command line interface.
type OptionSpec struct {
	Usage               string // Formatted usage string
	UnknownOptionsFatal bool   // Whether to die on unknown flags [true]
	UnknownValuesFatal  bool   // Whether to die on extra nonflags [false]

	ParseCallback func(*OptionSpec, string, *string) // Custom callback function
	Exit          func(code int)                     // Function to use for exiting [os.Exit]
	ErrorWriter   io.Writer                          // Alternate Writer for usage writing

	aliases     map[string]string
	defaults    map[string]string
	requiresArg map[string]bool
}

// SetUnknownOptionsFatal is a conveience function designed to be chained
// after NewOptions.
func (s *OptionSpec) SetUnknownOptionsFatal(val bool) *OptionSpec {
	s.UnknownOptionsFatal = val
	return s
}

// SetUnknownValuesFatal is a conveience function designed to be chained
// after NewOptions.
func (s *OptionSpec) SetUnknownValuesFatal(val bool) *OptionSpec {
	s.UnknownValuesFatal = val
	return s
}

// SetParseCallback is a convencience function designed to be chained after
// NewOptions.
func (s *OptionSpec) SetParseCallback(callback func(*OptionSpec, string, *string)) *OptionSpec {
	s.ParseCallback = callback
	return s
}

// NewOptions takes a string speficiation of a command line interface and
// returns an OptionSpec for you to call Parse on.
func NewOptions(spec string) *OptionSpec {
	// TODO(gaal): move to constant
	flagSpec := regexp.MustCompile(`^([-\w,]+)(=?)\s+(.*)$`)
	// Not folded into previous pattern because that would necessitate FindStringSubmatchIndex.
	defaultValue := regexp.MustCompile(`\[(.*)\]$`)

	s := &OptionSpec{UnknownOptionsFatal: true, UnknownValuesFatal: false, Exit: os.Exit}
	s.aliases = make(map[string]string)
	s.defaults = make(map[string]string)
	s.requiresArg = make(map[string]bool)
	stanza := 0 // synopsis
	specLines := strings.Split(spec, "\n")
	for n, l := range specLines {
		switch stanza {
		case 0:
			{
				if l == "--" {
					s.Usage += "\n"
					stanza++
					continue
				}
				s.Usage += l + "\n"
			}
		case 1:
			{
				if l == "" {
					s.Usage += "\n"
					continue
				}
				parts := flagSpec.FindStringSubmatch(l)
				if parts == nil {
					panic(fmt.Sprint(n, ": no parse: ", l))
				}
				names := strings.Split(parts[1], ",")
				canonical := names[len(names)-1]
				for _, name := range names {
					if _, dup := s.aliases[name]; dup {
						panic(fmt.Sprint(n, ": duplicate name: ", name))
					}
					if name == "" || name == "-" || name == "--" {
						panic(fmt.Sprint(n, ": bad name: ", name))
					}

					s.aliases[name] = canonical
				}
				if parts[2] == "=" {
					s.requiresArg[canonical] = true
				}
				if def := defaultValue.FindStringSubmatch(parts[3]); def != nil {
					s.defaults[canonical] = def[1]
				}
				// TODO(gaal): linewrap.
				s.Usage += "  " + strings.Join(smap(prettyFlag, names), ", ") +
					parts[2] + "  " + parts[3] + "\n"
			}
		default:
			panic(fmt.Sprint(n, ": no parse: ", spec))
		}
	}
	return s
}

// GetCanonical returns the canonical name of an option, or the empty string if
// the option is unkown. Useful to tidy up switch statements when using the
// custom callback interface.
func (s *OptionSpec) GetCanonical(option string) string {
	return s.aliases[option]
}

// BUG(gaal): Negated options ("--no-frobulate") are not yet supported.

// Parse performs the actual parsing of a command line according to an
// OptionSpec.
// It returns an Options value; see the package description for an overview
// of what it means and how to use it.
// In case of parse error, a panic is thrown.
// TODO(gaal): decide if gentler error signalling is more useful.
func (s *OptionSpec) Parse(args []string) Options {
	// TODO(gaal): extract to constant.
	flagRe := regexp.MustCompile(`^((--?)([-\w]+))(=(.*))?$`)

	opt := Options{
		opts:     make(map[string]string),
		Flags:    make([][]string, 0),
		Extra:    make([]string, 0),
		Leftover: make([]string, 0),
	}
	opt.opts = make(map[string]string)
	for flag, def := range s.defaults {
		opt.opts[flag] = def
	}
	opt.known = make(map[string]bool)
	for _, canonical := range s.aliases {
		opt.known[canonical] = true
	}

	for i := 0; i < len(args); i++ { // Can't use range because we may bump i.
		val := args[i]
		if val == "--" {
			opt.Leftover = append(opt.Leftover, args[i+1:]...)
			break
		}

		flagParts := flagRe.FindStringSubmatch(val)
		if flagParts == nil { // This is not a flag.
			if s.UnknownValuesFatal {
				panic("Unexpected argument: " + val + "\n" + s.Usage)
			}
			opt.Extra = append(opt.Extra, val)
			continue
		}
		presentedFlag := flagParts[1] // "presented" by the user.
		presentedDash := flagParts[2]
		presentedFlagName := flagParts[3]
		haveSelfValue := flagParts[4] != ""
		selfValue := flagParts[5]
		canonical, known := s.aliases[presentedFlagName]

		callback := s.ParseCallback
		if callback == nil {
			// The standard parse parse callback is in cahoots with the parser;
			// it closes on more state than user callbacks have access to. We
			// need to consider exposing more of it, e.g. both the presented
			// flag	(which is passed as a parameter now) and the canonical one
			// are interesting. But we don't want to complicate things too much,
			// so we'll probably not allow winding back an argument.
			callback = func(optionSpec *OptionSpec, option string, value *string) {
				unknown := func(k bool) bool {
					if !k && s.UnknownOptionsFatal {
						s.PrintUsageAndExit("Unkown option: " + option)
						return true // not reached
					}
					return !k
				}
				if presentedDash == "-" && len(presentedFlagName) > 1 { // Clustering, -abc
					for j, shortR := range presentedFlagName {
						short := string(shortR)
						isLast := j == len(presentedFlagName)-1
						canonicalC, knownC := s.aliases[short]
						if !unknown(knownC) {
							if s.requiresArg[canonicalC] {
								if value == nil || !isLast {
									s.PrintUsageAndExit("Missing argument: " + short)
									return // not reached
								}
								opt.opts[canonicalC] = *value
							} else {
								if value != nil && isLast {
									s.PrintUsageAndExit("Unexpected argument: " + short + ": " + *value)
									return // not reached
								}
								opt.opts[canonicalC] = fmt.Sprintf("%d", opt.GetInt(canonicalC)+1)
							}
						}
					}
				} else if !unknown(known) {
					if s.requiresArg[canonical] {
						if value == nil {
							s.PrintUsageAndExit("Missing argument: " + option)
							return // not reached
						}
						opt.opts[canonical] = *value
					} else {
						if value != nil {
							// Unlike the above nil check, reaching here is a programming bug.
							panic("Unexpected argument: " + option + ": " + *value)
						}
						opt.opts[canonical] = fmt.Sprintf("%d", opt.GetInt(canonical)+1)
					}
				}
				if value != nil {
					opt.Flags = append(opt.Flags, []string{presentedFlag, *value})
				} else {
					opt.Flags = append(opt.Flags, []string{presentedFlag})
				}
			}
		}

		maybeClustering := !known && presentedDash == "-"
		var nextArg *string = nil
		if i < len(args)-1 {
			nextArg = &(args[i+1])
		}
		needsArg := known && s.requiresArg[canonical]
		if !known && nextArg != nil && !strings.HasPrefix(*nextArg, "-") { // best effort unknown
			needsArg = true
		}
		if maybeClustering && haveSelfValue {
			needsArg = true
		}
		lastClustered := string(s.aliases[presentedFlagName[len(presentedFlagName)-1:]])
		if maybeClustering && s.requiresArg[lastClustered] {
			needsArg = true
		}
		arg := func() *string {
			if haveSelfValue {
				return &selfValue
			}
			i++
			return nextArg
		}
		if needsArg {
			callback(s, presentedFlagName, arg())
		} else {
			callback(s, presentedFlagName, nil)
		}

	}

	return opt
}

// PrintUsageAndExit writes the usage string and exits the program.
// If an error message is given, usage is written to standard error.
// Otherwise, it is written to standard output; this makes invocations
// such as "myprog --help | less" work as the user expects.
// Likewise, the status code is zero when no error was given.
func (s *OptionSpec) PrintUsageAndExit(err string) {
	printMsg := func(f io.Writer, format string, vs ...interface{}) {
		if s.ErrorWriter != nil {
			f = s.ErrorWriter
		}
		fmt.Fprintf(f, format, vs...)
	}
	if err == "" {
		printMsg(os.Stdout, "%s\n", s.Usage)
		s.Exit(0)
	}
	printMsg(os.Stderr, "%s\n%s\n", err, s.Usage)
	s.Exit(EX_USAGE)
}

func smap(f func(string) string, vs []string) []string {
	var out []string
	for _, v := range vs {
		out = append(out, f(v))
	}
	return out
}

func prettyFlag(flg string) string {
	if len(flg) == 1 {
		return "-" + flg
	}
	return "--" + flg
}
