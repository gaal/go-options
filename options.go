package options

import (
	"fmt"
	"regexp"
	"strings"
)

type Options struct {
	opts  map[string]string
	known map[string]bool
}

func (o *Options) Get(flag string) string {
	val, ok := o.opts[flag]
	if !ok {
		if !o.known[flag] {
			panic(fmt.Sprintf("[Programmer error] Unknown option: %s\ndump: %+v", flag, *o))
		}
	}
	return val
}

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

func (o *Options) GetBool(flag string) bool {
	val := o.Get(flag)
	if val == "" || val == "0" || val == "false" ||
		val == "off" || val == "nil" || val == "no" {
		return false
	}
	return true
}

type OptionSpec struct {
	Usage               string
	UnknownOptionsFatal bool // Whether to die on unknown flags [true]
	UnknownValuesFatal  bool // Whether to die un extra nonflags [false]
	aliases             map[string]string
	defaults            map[string]string
	short               map[string]bool // Single-char aliases, for clustering
	requiresArg         map[string]bool
}

func NewOptions(spec string) *OptionSpec {
	// TODO(gaal): move to constant
	flagSpec := regexp.MustCompile(`^([-\w,]+)(=?)\s+(.*)$`)
	// Not folded into previous pattern because that would necessitate FindStringSubmatchIndex.
	defaultValue := regexp.MustCompile(`\[(.*)\]$`)

	s := &OptionSpec{UnknownOptionsFatal: true, UnknownValuesFatal: false}
	s.aliases = make(map[string]string)
	s.defaults = make(map[string]string)
	s.short = make(map[string]bool)
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

					if len(name) == 1 {
						s.short[name] = true
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

func (s *OptionSpec) Parse(args []string) (Options, [][]string, []string) {
	// TODO(gaal): extract to constant.
	flagRe := regexp.MustCompile(`^((--?)([-\w]+))(=(.*))?$`)

	opt := Options{}
	opt.opts = make(map[string]string)
	for flag, def := range s.defaults {
		opt.opts[flag] = def
	}
	opt.known = make(map[string]bool)
	for _, canonical := range s.aliases {
		opt.known[canonical] = true
	}
	flags := make([][]string, 0)
	extra := make([]string, 0)

	for i := 0; i < len(args); i++ { // Can't use range because we may bump i.
		val := args[i]
		if val == "--" {
			break
		}

		flagParts := flagRe.FindStringSubmatch(val)
		if flagParts == nil { // This is not a flag.
			if s.UnknownValuesFatal {
				panic("Unexpected argument: " + val)
			}
			extra = append(extra, val)
			continue
		}
		presentedFlag := flagParts[1] // "presented" by the user.
		presentedFlagName := flagParts[3]
		haveSelfValue := flagParts[4] != ""
		selfValue := flagParts[5]
		canonical, haveCanonical := s.aliases[presentedFlagName]
		var nextArg *string = nil
		if i < len(args)-1 {
			nextArg = &(args[i+1])
		}

		recordOptionValue := func(value string) {
			if haveCanonical {
				opt.opts[canonical] = value
			}
			flags = append(flags, []string{presentedFlag, value})
		}
		recordOptionNoValue := func() {
			if haveCanonical {
				opt.opts[canonical] = fmt.Sprintf("%d", opt.GetInt(canonical)+1)
			}
			flags = append(flags, []string{presentedFlag})
		}

		if haveCanonical {
			if s.requiresArg[canonical] {
				if haveSelfValue {
					recordOptionValue(selfValue)
				} else if nextArg != nil {
					recordOptionValue(*nextArg)
					i++
				} else {
					panic("Option requires argument: " + canonical)
				}
			} else {
				// TODO(gaal): decide what to do: we were given an argument to
				// an option that doesn't take one. Do we treat this as an
				// optional argument and just record it? Panic?
				if haveSelfValue {
					panic("Option does not take argument: " + canonical)
				}
				recordOptionNoValue()
			}
		} else { // Unknown option: try to do the right thing.
			if s.UnknownOptionsFatal {
				panic("Unexpected option argument: " + val)
			}
			if haveSelfValue {
				recordOptionValue(selfValue)
			} else if nextArg != nil && !strings.HasPrefix(*nextArg, "-") {
				// Silently assume the next argument is an option value UNLESS
				// it syntactically looks like another flag. But note we don't
				// check the flag is known.
				recordOptionValue(*nextArg)
				i++
			} else {
				recordOptionNoValue()
			}
		}
	}

	return opt, flags, extra
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
