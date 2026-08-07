PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: all build install uninstall test clean

all: build

build:
	go build -o tskr .

install: build
	./tskr install --dir $(BINDIR)

uninstall:
	rm -f $(BINDIR)/tskr

test:
	go test -v ./...

clean:
	rm -f tskr
