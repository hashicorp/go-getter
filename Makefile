test:
	go get -d -v -t ./...
	go test -v ./...


package:
	go get github.com/mitchellh/gox
	gox -os="darwin linux windows" -arch="amd64" ./cmd/go-getter
	upx go-getter_*
	mv go-getter_darwin_amd64  go-getter_osx
	mv go-getter_linux_amd64  go-getter
	mv go-getter_windows_amd64.exe  go-getter.exe