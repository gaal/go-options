package options

import (
	"fmt"
	"testing"

	tu "github.com/gaal/go-util/testingutil"
)

func TestNewOptions_trivial(t *testing.T) {
	s := NewOptions("TestNewOptions_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	tu.ExpectEqual(t, s.aliases["ccc"], "ccc", "canonical name")
	tu.ExpectEqual(t, s.aliases["a"], "ccc", "alternate name")
	tu.ExpectEqual(t, s.aliases["bbb"], "ccc", "alternate name")
	tu.ExpectEqual(t, s.defaults["ccc"], "def", "default value")
	// This'll change (wrapping etc.) so it's not really worth testing too much.
	tu.ExpectEqual(
		t, s.Usage, "TestNewOptions_trivial\n\n  -a, --bbb, --ccc=  doc [def]\n",
		"usage string")
}

func TestParse_trivialDefault(t *testing.T) {
	s := NewOptions("TestParse_trivialDefault\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{})
	tu.ExpectEqual(t, opt.Get("ccc"), "def", "default (via canonical)")
	tu.ExpectEqual(t, opt.Flags, [][]string{}, "no flags specified")
	tu.ExpectEqual(t, opt.Extra, []string{}, "no extra args given")
}

func TestParse_trivial(t *testing.T) {
	s := NewOptions("TestParse_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	test := func(name string) {
		opt := s.Parse([]string{name, "myval"})
		tu.ExpectEqual(t, opt.opts["ccc"], "myval", "canonical direct access - "+name)
		tu.ExpectEqual(t, opt.Get("ccc"), "myval", "Get - "+name)
		tu.ExpectEqual(t, opt.Flags, [][]string{[]string{name, "myval"}}, "flags specified - "+name)
		tu.ExpectEqual(t, opt.Extra, []string{}, "no extra args given")
	}
	test("--ccc")
	test("--bbb")
	test("-a")
}

func TestParse_trivialSelfVal(t *testing.T) {
	s := NewOptions("TestParse_trivialSelfVal\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	test := func(name string) {
		opt := s.Parse([]string{name + "=myval"})
		tu.ExpectEqual(t, opt.opts["ccc"], "myval", "canonical direct access - "+name)
		tu.ExpectEqual(t, opt.Get("ccc"), "myval", "Get - "+name)
		tu.ExpectEqual(t, opt.Flags, [][]string{[]string{name, "myval"}}, "flags specified - "+name)
		tu.ExpectEqual(t, opt.Extra, []string{}, "no extra args given")
	}
	test("--ccc")
	test("--bbb")
	test("-a")
}

func TestParse_missingArgument(t *testing.T) {
	s := NewOptions("TestParse_missingArgument\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}
	tu.ExpectDie(t, func() { s.Parse([]string{"--ccc"}) }, "missing required param")
}

func TestParse_extra(t *testing.T) {
	s := NewOptions("TestParse_extra\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"extra1", "--ccc", "myval", "extra2", "extra3=foo"})
	tu.ExpectEqual(t, opt.Get("ccc"), "myval", "Get")
	tu.ExpectEqual(t, opt.Flags, [][]string{[]string{"--ccc", "myval"}}, "flags specified")
	tu.ExpectEqual(t, opt.Extra, []string{"extra1", "extra2", "extra3=foo"}, "extra args given")

	s.SetUnknownValuesFatal(true)
	tu.ExpectDie(t, func() {
		s.Parse([]string{"extra1", "--ccc", "myval", "extra2"})
	}, "dies on extras when asked to")
}

func TestParse_leftover(t *testing.T) {
	s := NewOptions("TestParse_leftover\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	s.SetUnknownValuesFatal(true)
	opt := s.Parse([]string{"--ccc", "myval"})
	tu.ExpectEqual(t, opt.Leftover, []string{}, "no leftover args given")

	opt = s.Parse([]string{"--ccc", "myval", "--"})
	tu.ExpectEqual(t, opt.Leftover, []string{}, "no leftover args given (with --)")

	opt = s.Parse([]string{"--ccc", "myval", "--", "leftover1", "leftover2"})
	tu.ExpectEqual(t, opt.Leftover, []string{"leftover1", "leftover2"}, "leftover args given")
}

func TestParse_unknownFlags(t *testing.T) {
	s := NewOptions("TestParse_unknownFlags\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}

	tu.ExpectDie(t, func() {
		s.Parse([]string{"--ccc", "myval", "--unk"})
	}, "dies on unknown options unless asked not to")

	s.SetUnknownOptionsFatal(false)
	opt := s.Parse([]string{"--unk1", "--ccc", "myval", "--unk2", "val2", "--unk3"})
	tu.ExpectEqual(t, opt.Get("ccc"), "myval", "Get")
	tu.ExpectEqual(t, opt.Flags, [][]string{
		[]string{"--unk1"},
		[]string{"--ccc", "myval"},
		[]string{"--unk2", "val2"},
		[]string{"--unk3"}},
		"flags specified")
	tu.ExpectEqual(t, opt.Extra, []string{}, "no extra args given")
}

func TestParse_override(t *testing.T) {
	s := NewOptions("TestParse_override\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"--bbb", "111", "--ccc", "222", "-a", "333"})
	tu.ExpectEqual(t, opt.Get("ccc"), "333", "last flag wins")
}

func TestParse_counting(t *testing.T) {
	s := NewOptions("TestParse_counting\n--\na,bbb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-a"})
	tu.ExpectEqual(t, opt.GetInt("ccc"), 1, "implicit value")

	opt = s.Parse([]string{"-a", "-a", "--ccc"})
	tu.ExpectEqual(t, opt.GetInt("ccc"), 3, "implicit value - repetitions")
}

func TestNewOptions_dupe(t *testing.T) {
	spec := `
--
a,bbb,ccc an option
d,bbb,eee an option with dupe`
	tu.ExpectDie(t, func() { NewOptions(spec) })
}

func TestGetAll(t *testing.T) {
	tu.ExpectEqual(
		t,
		GetAll("elk", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
		[]string{},
		"GetAll - nothing there")
	tu.ExpectEqual(
		t,
		GetAll("foo", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
		[]string{"aaa", "bbb"},
		"GetAll")
}

func TestCallbackInterface(t *testing.T) {
	s := NewOptions("TestCallbackInterface\n--\na,bbb,ccc= doc\nddd more doc\n")
	var (
		ccc     string
		ddd     bool
		unknown [][]string
	)
	s.ParseCallback = func(spec *OptionSpec, option string, argument *string) {
		if argument != nil {
			switch option {
			case "a", "bbb", "ccc":
				ccc = *argument
			default:
				unknown = append(unknown, []string{option, *argument})
			}
		} else {
			switch option {
			case "ddd":
				ddd = true
			default:
				unknown = append(unknown, []string{option})
			}
		}
	}
	opt := s.Parse(
		[]string{"--unk1", "--ccc", "myval", "--bbb=noooo", "hi", "a=b", "-a", "myotherval",
			"--unk2", "val2", "--ddd", "--unk3"})
	tu.ExpectEqual(t, ccc, "myotherval", "known option")
	tu.ExpectEqual(t, ddd, true, "known option")
	tu.ExpectEqual(
		t,
		unknown,
		[][]string{[]string{"unk1"}, []string{"unk2", "val2"}, []string{"unk3"}},
		"unknown options, with and without arguments")
	tu.ExpectEqual(t, opt.Extra, []string{"hi", "a=b"}, "extra")
}

func TestClustering_simple(t *testing.T) {
	s := NewOptions("TestClustering_simple\n--\na,bbb doc\nb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-abbb"})
	tu.ExpectEqual(t, opt.GetBool("bbb"), true, "clustering - simple")
	tu.ExpectEqual(t, opt.GetInt("bbb"), 1, "clustering - simple")
	tu.ExpectEqual(t, opt.GetInt("ccc"), 3, "clustering - increment")
}

func TestClustering_smoosh(t *testing.T) {
	s := NewOptions("TestClustering_smoosh\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab=foo"})
	tu.ExpectEqual(t, opt.GetInt("bbb"), 2, "clustering - smooshing")
	tu.ExpectEqual(t, opt.Get("ccc"), "foo", "clustering - smooshing")
}

func TestClustering_smooshWithSpace(t *testing.T) {
	s := NewOptions("TestClustering_smooshWithSpace\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab", "foo"})
	tu.ExpectEqual(t, opt.GetInt("bbb"), 2, "clustering - smooshing with a space")
	tu.ExpectEqual(t, opt.Get("ccc"), "foo", "clustering - smooshing with a space")

	opt = s.Parse([]string{"-aab", "-a"})
	tu.ExpectEqual(t, opt.GetInt("bbb"), 2, "clustering - smooshing with a space")
	tu.ExpectEqual(t, opt.Get("ccc"), "-a", "clustering - smooshing with a space")
}

func TestClustering_missingArg(t *testing.T) {
	s := NewOptions("TestClustering_missingArg\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}
	tu.ExpectDie(t, func() { s.Parse([]string{"-aab"}) })
}

func exitToPanic(code int) {
	panic(fmt.Sprintf("exiting with code: %d", code))
}

type devNull struct{}

func (d devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}
