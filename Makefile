MAKEFLAG += --warn-undefined-variables

BIN = monocol

PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

.PHONY: netboard

netboard:
	go build -trimpath .

default: netboard

install: netboard
	install -d $(BINDIR)
	install -s $(BIN) $(BINDIR)
