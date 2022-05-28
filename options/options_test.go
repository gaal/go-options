package options

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewOptions_trivial(t *testing.T) {
	s := NewOptions("TestNewOptions_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	wantAliases := map[string]string{
		"ccc": "ccc",
		"a":   "ccc",
		"bbb": "ccc",
	}
	if diff := cmp.Diff(wantAliases, s.aliases); diff != "" {
		t.Errorf("a,bbb,ccc= doc [def] resulted in wrong aliases (-want,+got):\n%s", diff)
	}
	if got, want := s.defaults["ccc"], "def"; got != want {
		t.Errorf("a,bbb,ccc= doc [def] resulted in wrong default=%q, want=%q", got, want)
	}
}

func TestParse_trivialDefault(t *testing.T) {
	s := NewOptions("TestParse_trivialDefault\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{})
	if got := opt.Get("ccc"); got != "def" {
		t.Errorf(`opt.Get("ccc")=%q, want=%q`, got, "def")
	}
	if len(opt.Flags) > 0 {
		t.Errorf("unexpected flags parsed: %q", opt.Flags)
	}
	if len(opt.Extra) > 0 {
		t.Errorf("unexpected extras parsed: %q", opt.Extra)
	}
}

func TestParse_trivial(t *testing.T) {
	s := NewOptions("TestParse_trivial\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	test := func(name string) {
		opt := s.Parse([]string{name, "myval"})
		if got, want := opt.opts["ccc"], "myval"; got != want {
			t.Errorf("%s: canonical direct access=%q, want=%q", name, got, want)
		}
		if got, want := opt.Get("ccc"), "myval"; got != want {
			t.Errorf("%s: Get=%q, want=%q", name, got, want)
		}
		if diff := cmp.Diff(opt.Flags, [][]string{[]string{name, "myval"}}); diff != "" {
			t.Errorf("%s: flags differ (-want+got):\n%s", name, diff)
		}
		if len(opt.Extra) > 0 {
			t.Errorf("unexpected extras parsed: %q", opt.Extra)
		}
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
		if got, want := opt.opts["ccc"], "myval"; got != want {
			t.Errorf("%s: canonical direct access=%q, want=%q", name, got, want)
		}
		if got, want := opt.Get("ccc"), "myval"; got != want {
			t.Errorf("%s: Get=%q, want=%q", name, got, want)
		}
		if diff := cmp.Diff(opt.Flags, [][]string{[]string{name, "myval"}}); diff != "" {
			t.Errorf("%s: flags differ (-want+got):\n%s", name, diff)
		}
		if len(opt.Extra) > 0 {
			t.Errorf("unexpected extras parsed: %q", opt.Extra)
		}
	}
	test("--ccc")
	test("--bbb")
	test("-a")
}

func TestParse_missingArgument(t *testing.T) {
	s := NewOptions("TestParse_missingArgument\n--\na,bbb,ccc= doc [def]")
	var i int
	s.Exit = func(code int) { i = code }
	s.ErrorWriter = devNull{}

	s.Parse([]string{"--ccc"})
	if i == 0 {
		t.Errorf("expected failure with nonzero code, got=0")
	}
}

func TestParse_extra(t *testing.T) {
	s := NewOptions("TestParse_extra\n--\nccc= doc [def]")
	// var i int
	// s.Exit = func(code int) { i = code }
	s.Exit = exitToPanic
	opt := s.Parse([]string{"extra1", "--ccc", "myval", "extra2", "extra3=foo"})
	if got, want := opt.Get("ccc"), "myval"; got != want {
		t.Errorf(`opt.Get("ccc")=%q, want=%q`, got, want)
	}
	if diff := cmp.Diff(opt.Flags, [][]string{[]string{"--ccc", "myval"}}); diff != "" {
		t.Errorf("flags diff (-want+got):\n%s,", diff)
	}
	if diff := cmp.Diff(opt.Extra, []string{"extra1", "extra2", "extra3=foo"}); diff != "" {
		t.Errorf("extra diff (-want+got):\n%s,", diff)
	}

	// TODO(gaal): fix death test
	/*
		s.SetUnknownValuesFatal(true)
		s.Parse([]string{"extra1", "--ccc", "myval", "extra2"})
		if i == 0 {
			t.Errorf("expected failure on extras when asked to")
		}
	*/
}

func TestParse_leftover(t *testing.T) {
	s := NewOptions("TestParse_leftover\n--\nccc= doc [def]")
	s.Exit = exitToPanic
	s.SetUnknownValuesFatal(true)
	opt := s.Parse([]string{"--ccc", "myval"})
	if len(opt.Leftover) > 0 {
		t.Errorf("leftover args: %q", opt.Leftover)
	}

	opt = s.Parse([]string{"--ccc", "myval", "--"})
	if len(opt.Leftover) > 0 {
		t.Errorf("leftover args (with --): %q", opt.Leftover)
	}

	opt = s.Parse([]string{"--ccc", "myval", "--", "leftover1", "leftover2"})
	if diff := cmp.Diff(opt.Leftover, []string{"leftover1", "leftover2"}); diff != "" {
		t.Errorf("leftover args diff (-want+got):\n%s", diff)
	}
}

func TestParse_unknownFlags(t *testing.T) {
	s := NewOptions("TestParse_unknownFlags\n--\nccc= doc [def]")
	var i int
	s.Exit = func(code int) { i = code }
	s.ErrorWriter = devNull{}

	s.Parse([]string{"--ccc", "myval", "--unk"})
	if i == 0 {
		t.Fatalf(`Parse([]string{"--ccc", "myval", "--unk"}) unexpectedly passed`)
	}
	i = 0

	s.SetUnknownOptionsFatal(false)
	opt := s.Parse([]string{"--unk1", "--ccc", "myval", "--unk2", "val2", "--unk3"})
	if got, want := opt.Get("ccc"), "myval"; got != want {
		t.Errorf(`opt.Get("ccc")=%q, want=%q`, got, want)
	}
	want := [][]string{
		[]string{"--unk1"},
		[]string{"--ccc", "myval"},
		[]string{"--unk2", "val2"},
		[]string{"--unk3"},
	}
	if diff := cmp.Diff(opt.Flags, want); diff != "" {
		t.Errorf("opt.Flags diff (-want+got):\n%s", diff)
	}
	if len(opt.Extra) > 0 {
		t.Errorf("extra args: %q", opt.Extra)
	}
}

func TestParse_override(t *testing.T) {
	s := NewOptions("TestParse_override\n--\na,bbb,ccc= doc [def]")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"--bbb", "111", "--ccc", "222", "-a", "333"})
	if got, want := opt.Get("ccc"), "333"; got != want {
		t.Errorf(`opt.Get("ccc")=%q, want=%q`, got, want)
	}
}

func TestParse_counting(t *testing.T) {
	s := NewOptions("TestParse_counting\n--\na,bbb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-a"})
	if got, want := opt.GetInt("ccc"), 1; got != want {
		t.Errorf(`opt.GetInt("ccc")=%d, want=%d`, got, want)
	}

	opt = s.Parse([]string{"-a", "-a", "--ccc"})
	if got, want := opt.GetInt("ccc"), 3; got != want {
		t.Errorf(`opt.GetInt("ccc") implicit value - repetitions=%d, want=%d`, got, want)
	}
}

func TestNewOptions_dupe(t *testing.T) {
	// TODO(gaal): cover.
	_ = `
--
a,bbb,ccc an option
d,bbb,eee an option with dupe`
	// tu.ExpectDie(t, func() { NewOptions(spec) })
}

func TestGetAll(t *testing.T) {
	if diff := cmp.Diff(
		GetAll("elk", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
		[]string{}); diff != "" {
		t.Errorf("GetAll - nothing there diff (-want+got):%s", diff)
	}

	if diff := cmp.Diff(
		GetAll("foo", [][]string{[]string{"foo", "aaa"}, []string{"bar"}, []string{"foo", "bbb"}}),
		[]string{"aaa", "bbb"}); diff != "" {
		t.Errorf("GetAll diff (-want+got):%s", diff)
	}
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
	if got, want := ccc, "myotherval"; got != want {
		t.Errorf("known option = %q, want = %q", got, want)
	}
	if got, want := ddd, true; got != want {
		t.Errorf("known option = %t, want = %t", got, want)
	}
	if diff := cmp.Diff(
		unknown,
		[][]string{[]string{"unk1"}, []string{"unk2", "val2"}, []string{"unk3"}}); diff != "" {
		t.Errorf("unknown options, with and without arguments diff (-want+got):\n%s", diff)
	}
	if diff := cmp.Diff(opt.Extra, []string{"hi", "a=b"}); diff != "" {
		t.Errorf("extra diff (-want+got):\n%s", diff)
	}
}

func TestClustering_simple(t *testing.T) {
	s := NewOptions("TestClustering_simple\n--\na,bbb doc\nb,ccc doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-abbb"})
	if got, want := opt.GetBool("bbb"), true; got != want {
		t.Errorf(`clustering - simple = %t, want = %t`, got, want)
	}
	if got, want := opt.GetInt("bbb"), 1; got != want {
		t.Errorf(`clustering - simple = %q, want = %q`, got, want)
	}
	if got, want := opt.GetInt("ccc"), 3; got != want {
		t.Errorf(`clustering - increment = %q, want = %q`, got, want)
	}
}

func TestClustering_smoosh(t *testing.T) {
	s := NewOptions("TestClustering_smoosh\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab=foo"})
	if got, want := opt.GetInt("bbb"), 2; got != want {
		t.Errorf(`clustering - smooshing = %q, want = %q`, got, want)
	}
	if got, want := opt.Get("ccc"), "foo"; got != want {
		t.Errorf(`clustering - smooshing = %q, want = %q`, got, want)
	}
}

func TestClustering_smooshWithSpace(t *testing.T) {
	s := NewOptions("TestClustering_smooshWithSpace\n--\na,bbb doc\nb,ccc= doc")
	s.Exit = exitToPanic
	opt := s.Parse([]string{"-aab", "foo"})
	if got, want := opt.GetInt("bbb"), 2; got != want {
		t.Errorf(`clustering - smooshing with a space = %q, want = %q`, got, want)
	}
	if got, want := opt.Get("ccc"), "foo"; got != want {
		t.Errorf(`clustering - smooshing with a space = %q, want = %q`, got, want)
	}

	opt = s.Parse([]string{"-aab", "-a"})
	if got, want := opt.GetInt("bbb"), 2; got != want {
		t.Errorf(`clustering - smooshing with a space = %q, want = %q`, got, want)
	}
	if got, want := opt.Get("ccc"), "-a"; got != want {
		t.Errorf(`clustering - smooshing with a space = %q, want = %q`, got, want)
	}
}

func TestClustering_missingArg(t *testing.T) {
	s := NewOptions("TestClustering_missingArg\n--\na,bbb doc\nb,ccc= doc")
	var i int
	s.Exit = func(code int) { i = code }
	s.ErrorWriter = devNull{}
	s.Parse([]string{"-aab"})
	if i == 0 {
		t.Errorf("expected failure with nonzero code, got=0")
	}

}

func exitToPanic(code int) {
	panic(fmt.Sprintf("exiting with code: %d", code))
}

type devNull struct{}

func (d devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}
