include $(GOROOT)/src/Make.inc

all:
	cd pkg && gomake

install:
	cd pkg && gomake install

test:
	cd pkg && gomake test

example: install
	cd example && gomake example

clean:
	cd pkg && gomake clean
	cd example && gomake clean
