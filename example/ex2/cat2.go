// Example use of the options library, using callback/static style.
// Try running this with various options, including invalid ones.
package main

import (
	"fmt"
	"os"

	"github.com/gaal/go-options"
)

var (
	n, e, v bool
	in, out string
	r       int
)

const mySpec = `
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
`

func argCb(spec OptionSpec, option string, argument string) {
	switch option {
	case "input-encoding":
		in = argument
	case "output-encoding":
		out = argument
	case "repeat":
		fmt.Scanf(argument, "%d", &r)
	default:
		spec.PrintUsageAndExit("Unknown option: " + option)
	}
}

func noArgCb(spec OptionSpec, option string) {
	switch option {
	case "number":
		n = true
	case "escape":
		e = true
	case "verbose":
		v = true
	case "help":
		spec.PrintUsageAndExit("") // No error
	default:
		spec.PrintUsageAndExit("Unknown option: " + option)
	}
}

func main() {
	spec := options.NewOptions(mySpec).SetCallbacks(argCb, noArgCb)
	opt, flags, extra := spec.Parse(os.Args[1:])

	fmt.Printf("I will concatenate the files: %q\n", extra)
	if opt.GetBool("number") {
		fmt.Println("I will number each line")
	}
	if opt.GetBool("escape") {
		fmt.Println("I will escape each line")
	}
	if r := opt.GetInt("repeat"); r != 1 {
		fmt.Printf("I will repeat each line %d times\n", r)
	}
	if v := opt.GetInt("verbose"); v > 0 {
		fmt.Printf("I will be verbose (level %d)\n", v)
	}
	fmt.Printf("Input charset: %s\n", opt.Get("input-encoding"))
	fmt.Printf("Output charset: %s\n", opt.Get("output-encoding"))

	fmt.Printf("For reference, here are the flags you gave me: %v\n", flags)
}
