// Example use of the options library, using callback/static style.
// Try running this with various options, including invalid ones.
package main

import (
	"math/big"
	"fmt"
	"os"
	"crypto/rand"

	"github.com/gaal/go-options/options"
)

// Note how with callbacks, the programmer is responsible for default values.
var (
	n, e    bool
	in, out string = "utf-8", "utf-8"
	r, v    int    = 1, 0
	c       float32 = 0.1
	cInt    int64
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
c,cookie-chance=      chance to get a fortune cookie [0.1]
`

// Example custom option callback
func MyParseCallback(spec *options.OptionSpec, option string, argument *string) {
	if argument != nil {
		switch spec.GetCanonical(option) {
		case "input-encoding":
			in = *argument
		case "output-encoding":
			out = *argument
		case "repeat":
			fmt.Sscanf(*argument, "%d", &r)
		case "cookie-chance":
			fmt.Sscanf(*argument, "%f", &c)
			cInt = int64(c * 1000)
		default:
			spec.PrintUsageAndExit("Unknown option: " + option)
		}
	} else {
		switch spec.GetCanonical(option) {
		case "number":
			n = true
		case "escape":
			e = true
		case "verbose":
			v++
		default:
			if option == "help" {
				spec.PrintUsageAndExit("") // No error
			} else {
				spec.PrintUsageAndExit("Unknown option: " + option)
			}
		}
	}
}

func main() {
	spec := options.NewOptions(mySpec).SetParseCallback(MyParseCallback)
	opt := spec.Parse(os.Args[1:])

	fmt.Printf("I will concatenate the files: %q\n", opt.Extra)
	if n {
		fmt.Println("I will number each line")
	}
	if e {
		fmt.Println("I will escape each line")
	}
	if r != 1 {
		fmt.Printf("I will repeat each line %d times\n", r)
	}
	if v > 0 {
		fmt.Printf("I will be verbose (level %d)\n", v)
	}
	fmt.Printf("Input charset: %s\n", in)
	fmt.Printf("Output charset: %s\n", out)

	fmt.Printf("Chance for a cookie: %.2f\n", c)
	var max = big.NewInt(1000)
	rnd, _ := rand.Int(rand.Reader, max)
	if cInt > rnd.Int64() {
		fmt.Println("*** You got a cookie! Yay! ***")
	}
}
