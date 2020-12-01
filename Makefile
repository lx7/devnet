APP := dnd
#PLATFORMS := linux/arm linux/amd64
PLATFORMS := linux/amd64
DISTDIR = ./bin
CMDDIR = ./cmd/'$(APP)'

BROWSERCMD := /usr/bin/firefox 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

all: release

run: 
	go run . -c ../../configs/server.yaml

test:
	go test ./...

release: test $(PLATFORMS)

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

.PHONY: run release $(PLATFORMS) clean
