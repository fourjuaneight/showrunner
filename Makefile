VERSION := 0.1.0

.PHONY: run build install

run: showrunner.go go.mod
	chmod +x showrunner.go
	cd "$GOPATH" && go get github.com/erning/gorun

# https://golang.org/cmd/link/
build: showrunner.go go.mod
	sed -i '.bak' 's/BuildVersion string = ".*"/BuildVersion string = "${VERSION}"/g' showrunner.go
	go build -ldflags="-X 'main.BuildVersion=$(VERSION)'" ./showrunner.go

install: showrunner.go go.mod
	go install ./showrunner.go
