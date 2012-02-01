include $(GOROOT)/src/Make.inc

all:
	cd options && gomake

install:
	cd options && gomake install

test:
	cd options && gomake test

example: install
	cd example && gomake example

clean:
	cd options && gomake clean
	cd example && gomake clean
