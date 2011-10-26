# figure out what GOROOT is supposed to be
GOROOT ?= $(shell printf 't:;@echo $$(GOROOT)\n' | gomake -f -)
include $(GOROOT)/src/Make.inc

TARG=github.com/bmizerany/pq.go/buffer
GOFILES=\
	buffer.go\

include $(GOROOT)/src/Make.pkg
