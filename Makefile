.PHONY: install deps

install: deps assets.go
	go build -o statusbar
	sudo mv statusbar /usr/local/bin/statusbar
	if [ -f statusbar.json ]; then sudo cp statusbar.json /usr/local/etc/statusbar.json; fi

assets.go:
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -nomemcopy -o assets.go xbm/

deps:
	@$(call installed,go)
	@$(call installed,dzen2)
	@$(call installed,setxkbmap)

# checks whether a command is installed
define installed =
command -v $(1) >/dev/null 2>&1 || (echo "$(1) needs to be installed and available in your PATH"; exit 1)
endef
