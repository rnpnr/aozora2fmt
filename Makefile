# See LICENSE for license details.
PREFIX = /usr/local

LDFLAGS = -s

BINS = aozora2fmt

default: $(BINS)

aozora2fmt: aozora2fmt.go
	go build -ldflags "$(LDFLAGS)" $@.go

install: $(BINS)
	mkdir -p $(PREFIX)/bin
	cp $(BINS) $(PREFIX)/bin
	chmod 755 $(BINS:%=$(PREFIX)/bin/%)

uninstall:
	rm $(BINS:%=$(PREFIX)/bin/%)

clean:
	rm $(BINS)
