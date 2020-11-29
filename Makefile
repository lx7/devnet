APP := dnd
#PLATFORMS := linux/arm linux/amd64
PLATFORMS := linux/amd64
DISTDIR = ./bin
CMDDIR = ./cmd/'$(APP)'

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

clean:
	rm -r ${DISTDIR}

.PHONY: run release $(PLATFORMS) clean
