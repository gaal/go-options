go-options - command line parsing library for Go
================================================

To install: go get github.com/gaal/go-options/options

* Easy to use - no boilerplate
* Self-documenting - the spec turns into the usage string
* Powerful - doesn't do everything you ever dreamed of, but comes close
* Flexible - if you want control over parsing, provide your own callback.

This design is inspired by `git rev-parse --parseopt` and the discussion of
`bup.options` here: <http://apenwarr.ca/log/?m=201111#02>. There are some
minor deviations.

* On the code side, you must access the opt structure with canonical option
names only. This is intended to reduce programmer errors. (This applies to
the simple, non-callback interface.)
* When I support negated options, I will not support unnegated aliases
for them as that can lead to more confusion than I deem worth harboring.

This package is distributed under the MIT/X license.

Comments? Please write me at <gaal@forum2.org>.
