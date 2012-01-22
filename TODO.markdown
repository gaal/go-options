* Simple clustering of short opts, e.g., `-abc`
* Smooshing with short opts, e.g., `-abcfoo` ==> `-a -b -c foo` magically
* Negated args, e.g., `--no-foo`. (Note, I'm not sure "no" needs to be
allowed as part of the name and even if so, all aliases should have the
same negation value. Probably just allow it on the user side.)
