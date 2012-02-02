// Example use of the options library, using opt/flags/extra style.
// Try running this with various options, including invalid ones.
package main

import (
    "fmt"
    "os"

    "github.com/gaal/go-options/options"
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
author=               authors you like (may be repeated)
`

func main() {
    spec := options.NewOptions(mySpec)
    opt := spec.Parse(os.Args[1:])

    fmt.Printf("I will concatenate the files: %q\n", opt.Extra)
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
	authors := options.GetAll("--author", opt.Flags)  // Note, you need "--".
	if len(authors) > 0 {
		fmt.Printf("You like these authors. I'll tell you if I see them: %q\n", authors)
	}

    fmt.Printf("For reference, here are the flags you gave me: %v\n", opt.Flags)
}
