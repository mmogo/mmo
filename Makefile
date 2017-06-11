SOURCEDIR=.
CLIENTDIR=$(SOURCEDIR)/client
SERVERDIR=$(SOURCEDIR)/server
PATCHERDIR=$(SOURCEDIR)/patcher
ASSETDIR=$(CLIENTDIR)/assets
ASSETS := $(shell find $(SOURCEDIR)/client/assets -name assets.go -prune -o -print)
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: all

all: bin/patcher bin/server bin/client.so

bin/patcher: $(SOURCES)
	pushd $(PATCHERDIR)
	go build -o ../bin/patcher main.go
	popd

bin/server: $(SOURCES)
	pushd $(SERVERDIR)
	go build -o ../bin/server main.go
	popd

bin/client.so: $(SOURCES)
	pushd $(CLIENTDIR)
	go build -buildmode=plugin -o ../bin/client.so main.go
	popd

$(ASSETDIR)/assets.go: $(ASSETS)
	pushd $(CLIENTDIR)
	go-bindata -o assets/assets.go -pkg assets -prefix assets/ assets/
	popd

.PHONY: clean
	rm -rf bin
	rm $(ASSETDIR)/assets.go