.PHONY: install deps

install:
	go-bindata -nomemcopy -o assets.go xbm/
	go build -o statusbar
	sudo mv statusbar /usr/local/bin/statusbar
	if [ -f statusbar.json ]; then sudo cp statusbar.json /usr/local/etc/statusbar.json; fi

deps:
	go get -u github.com/jteeuwen/go-bindata/...
