APP := devnet
PLATFORMS := linux/amd64
DISTDIR = ./bin
CMDDIR = ./cmd/'$(APP)'

BROWSERCMD := /usr/bin/firefox 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

all: release

generate:
	go generate ./...
	protoc --go_out . --go_opt=paths=source_relative proto/*.proto

check:
	go vet ./...
	errcheck ./...
	golint ./...
	staticcheck ./...

test:
	go test ./...

release: generate check test $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o '$(DISTDIR)/$(APP)-$(os)-$(arch)' '$(CMDDIR)'

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	$(BROWSERCMD) coverage.html

clean:
	rm -rf ${DISTDIR}
	find -H . -type f -name "coverage.out" -delete
	find -H . -type f -name "coverage.html" -delete
	find -H . -type f -name "*.pb.go" -delete
	find -H . -type f -name "*.gen.go" -delete
	go mod tidy

deps:
	$(info === Installing dependencies. This may take a while. ===)
	go get honnef.co/go/tools/cmd/staticcheck
	go get github.com/gotk3/gotk3/gtk
	go get github.com/mjibson/esc

.PHONY: run release $(PLATFORMS) cover clean deps
