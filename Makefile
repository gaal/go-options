include $(GOROOT)/src/Make.inc

install:
	cd pkg && gomake install

test:
	cd pkg && gomake test

example:
	cd pkg && gomake
	cd example && gomake example

clean:
	cd pkg && gomake clean
	cd example && gomake clean
