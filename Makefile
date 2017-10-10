prefix=/usr/local
PWD := $(shell pwd)
GOPATH :=$(PWD)/deps
appname = meteo

all:
	GOPATH=$(GOPATH) go get -d github.com/ugjka/$(appname)GOPATH
	GOPATH=$(GOPATH) go build -v -tags "gtk_3_20" -ldflags="-X main.icon=$(prefix)/share/icons/hicolor/256x256/apps/$(appname).png"

install:
	install -Dm755 $(appname) $(prefix)/bin/$(appname)
	install -Dm644 LICENSE "$(prefix)/share/licenses/$(appname)/LICENSE"
	install -Dm644 logo.png "$(prefix)/share/icons/hicolor/256x256/apps/$(appname).png"
	install -Dm644 $(appname).desktop "$(prefix)/share/applications/$(appname).desktop"

uninstall:
	rm "$(prefix)/bin/$(appname)"
	rm "$(prefix)/share/licenses/$(appname)/LICENSE"
	rm "$(prefix)/share/icons/hicolor/256x256/apps/$(appname).png"
	rm "$(prefix)/share/applications/$(appname).desktop"

clean:
	rm -rf $(GOPATH)
	rm $(appname)