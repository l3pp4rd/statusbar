.PHONY: install deps

install:
	go-bindata -nomemcopy -o assets.go xbm/
	go build -o statusbar
	sudo mv statusbar /usr/local/bin/statusbar

deps:
	go get -u github.com/jteeuwen/go-bindata/...
