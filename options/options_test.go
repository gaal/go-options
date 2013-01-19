package options

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestNewOptions_trivial(t *testing.T) {
	s := NewOptions("TestNewOptions_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	ExpectEquals(t, "ccc", s.aliases["ccc"], "canonical name")
	ExpectEquals(t, "ccc", s.aliases["a"], "alternate name")
	ExpectEquals(t, "ccc", s.aliases["bbb"], "alternate name")
	ExpectEquals(t, "def", s.defaults["ccc"], "default value")
	// This'll change (wrapping etc.) so it's not really worth testing too much.
	ExpectEquals(t, "TestNewOptions_trivial\n\n  -a, --bbb, --ccc=  doc [def]\n",
		s.Usage, "usage string")
}

func TestParse_trivialDefault(t *testing.T) {
	s := NewOptions("TestParse_trivialDefault\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{})
	ExpectEquals(t, "def", opt.Get("ccc"), "default (via canonical)")
	ExpectEquals(t, [][]string{}, opt.Flags, "no flags specified")
	ExpectEquals(t, []string{}, opt.Extra, "no extra args given")
}

func TestParse_trivial(t *testing.T) {
	s := NewOptions("TestParse_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	test := func(name string) {
		opt := s.Parse([]string{name, "myval"})
		ExpectEquals(t, "myval", opt.opts["ccc"], "canonical direct access - "+name)
		ExpectEquals(t, "myval", opt.Get("ccc"), "Get - "+name)
		ExpectEquals(t, [][]string{[]string{name, "myval"}}, opt.Flags, "flags specified - "+name)
		ExpectEquals(t, []string{}, opt.Extra, "no extra args given")
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
		ExpectEquals(t, "myval", opt.opts["ccc"], "canonical direct access - "+name)
		ExpectEquals(t, "myval", opt.Get("ccc"), "Get - "+name)
		ExpectEquals(t, [][]string{[]string{name, "myval"}}, opt.Flags, "flags specified - "+name)
		ExpectEquals(t, []string{}, opt.Extra, "no extra args given")
	}
	test("--ccc")
	test("--bbb")
	test("-a")
}

func TestParse_missingArgument(t *testing.T) {
	s := NewOptions("TestParse_missingArgument\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}
	ExpectDies(t, func() { s.Parse([]string{"--ccc"}) }, "missing required param")
}

func TestParse_extra(t *testing.T) {
	s := NewOptions("TestParse_extra\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"extra1", "--ccc", "myval", "extra2", "extra3=foo"})
	ExpectEquals(t, "myval", opt.Get("ccc"), "Get")
	ExpectEquals(t, [][]string{[]string{"--ccc", "myval"}}, opt.Flags, "flags specified")
	ExpectEquals(t, []string{"extra1", "extra2", "extra3=foo"}, opt.Extra, "extra args given")

	s.SetUnknownValuesFatal(true)
	ExpectDies(t, func() {
		s.Parse([]string{"extra1", "--ccc", "myval", "extra2"})
	}, "dies on extras when asked to")
}

func TestParse_leftover(t *testing.T) {
	s := NewOptions("TestParse_leftover\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	s.SetUnknownValuesFatal(true)
	opt := s.Parse([]string{"--ccc", "myval"})
	ExpectEquals(t, []string{}, opt.Leftover, "no leftover args given")

	opt = s.Parse([]string{"--ccc", "myval", "--"})
	ExpectEquals(t, []string{}, opt.Leftover, "no leftover args given (with --)")

	opt = s.Parse([]string{"--ccc", "myval", "--", "leftover1", "leftover2"})
	ExpectEquals(t, []string{"leftover1", "leftover2"}, opt.Leftover, "leftover args given")
}

func TestParse_unknownFlags(t *testing.T) {
	s := NewOptions("TestParse_unknownFlags\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}

	ExpectDies(t, func() {
		s.Parse([]string{"--ccc", "myval", "--unk"})
	}, "dies on unknown options unless asked not to")

	s.SetUnknownOptionsFatal(false)
	opt := s.Parse([]string{"--unk1", "--ccc", "myval", "--unk2", "val2", "--unk3"})
	ExpectEquals(t, "myval", opt.Get("ccc"), "Get")
	ExpectEquals(t, [][]string{
		[]string{"--unk1"},
		[]string{"--ccc", "myval"},
		[]string{"--unk2", "val2"},
		[]string{"--unk3"}},
		opt.Flags, "flags specified")
	ExpectEquals(t, []string{}, opt.Extra, "no extra args given")
}

func TestParse_override(t *testing.T) {
	s := NewOptions("TestParse_override\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"--bbb", "111", "--ccc", "222", "-a", "333"})
	ExpectEquals(t, "333", opt.Get("ccc"), "last flag wins")
}

func TestParse_counting(t *testing.T) {
	s := NewOptions("TestParse_counting\n--\na,bbb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-a"})
	ExpectEquals(t, 1, opt.GetInt("ccc"), "implicit value")

	opt = s.Parse([]string{"-a", "-a", "--ccc"})
	ExpectEquals(t, 3, opt.GetInt("ccc"), "implicit value - repetitions")
}

func TestNewOptions_dupe(t *testing.T) {
	spec := `
--
a,bbb,ccc an option
d,bbb,eee an option with dupe`
	ExpectDies(t, func() { NewOptions(spec) })
}

func TestGetAll(t *testing.T) {
	ExpectEquals(
		t,
		[]string{},
		GetAll("elk", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
		"GetAll - nothing there")
	ExpectEquals(
		t,
		[]string{"aaa", "bbb"},
		GetAll("foo", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
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
	ExpectEquals(t, "myotherval", ccc, "known option")
	ExpectEquals(t, true, ddd, "known option")
	ExpectEquals(
		t,
		[][]string{[]string{"unk1"}, []string{"unk2", "val2"}, []string{"unk3"}},
		unknown,
		"unknown options, with and without arguments")
	ExpectEquals(t, []string{"hi", "a=b"}, opt.Extra, "extra")
}

func TestClustering_simple(t *testing.T) {
	s := NewOptions("TestClustering_simple\n--\na,bbb doc\nb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-abbb"})
	ExpectEquals(t, true, opt.GetBool("bbb"), "clustering - simple")
	ExpectEquals(t, 1, opt.GetInt("bbb"), "clustering - simple")
	ExpectEquals(t, 3, opt.GetInt("ccc"), "clustering - increment")
}

func TestClustering_smoosh(t *testing.T) {
	s := NewOptions("TestClustering_smoosh\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab=foo"})
	ExpectEquals(t, 2, opt.GetInt("bbb"), "clustering - smooshing")
	ExpectEquals(t, "foo", opt.Get("ccc"), "clustering - smooshing")
}

func TestClustering_smooshWithSpace(t *testing.T) {
	s := NewOptions("TestClustering_smooshWithSpace\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab", "foo"})
	ExpectEquals(t, 2, opt.GetInt("bbb"), "clustering - smooshing with a space")
	ExpectEquals(t, "foo", opt.Get("ccc"), "clustering - smooshing with a space")

	opt = s.Parse([]string{"-aab", "-a"})
	ExpectEquals(t, 2, opt.GetInt("bbb"), "clustering - smooshing with a space")
	ExpectEquals(t, "-a", opt.Get("ccc"), "clustering - smooshing with a space")
}

func TestClustering_missingArg(t *testing.T) {
	s := NewOptions("TestClustering_missingArg\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	s.ErrorWriter = devNull{}
	ExpectDies(t, func() { s.Parse([]string{"-aab"}) })
}

func exitToPanic(code int) {
	panic(fmt.Sprintf("exiting with code: %d", code))
}

type devNull struct{}

func (d devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// These are little testing utilities that I like. May move to a separate module one day.

func ExpectEquals(t *testing.T, expected, actual interface{}, desc ...string) {
	if !reflect.DeepEqual(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		desc1 := fmt.Sprintf("%s:%d", file, line)
		if len(desc) > 0 {
			desc1 += " " + fmt.Sprint(desc)
		}
		t.Errorf("%s\nExpected: %#v\nActual:   %#v\n", desc1, expected, actual)
	}
}

func ExpectDies(t *testing.T, f func(), desc ...string) {
	_, file, line, _ := runtime.Caller(1)
	defer func() {
		if x := recover(); x == nil {
			t.Errorf("%v\nExpected panic:%s:%d\n", desc, file, line)
		}
	}()
	f()
}

func TestExpectDies(t *testing.T) {
	ExpectDies(t, func() { panic("aaaaahh") }, "simple panic dies")
	t1 := new(testing.T)
	ExpectDies(t1, func() {}, "doesn't die")
	ExpectEquals(t, true, t1.Failed(), "ExpectDies on something that doesn't die fails")
}
