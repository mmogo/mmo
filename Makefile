SOURCEDIR=.
SHAREDDIR=$(SOURCEDIR)/shared
CLIENTDIR=$(SOURCEDIR)/client
SERVERDIR=$(SOURCEDIR)/server
PATCHERDIR=$(SOURCEDIR)/patcher
ASSETDIR=$(CLIENTDIR)/assets
ASSETS := $(shell find $(SOURCEDIR)/client/assets -name assets.go -prune -o -print)
OUTPUTDIR := $(SOURCEDIR)/bin

CLIENTSOURCES := $(shell find $(CLIENTDIR) $(SHAREDDIR) -name '*.go')
SERVERSOURCES := $(shell find $(SERVERDIR) $(SHAREDDIR) -name '*.go')
PATCHERSOURCES := $(shell find $(PATCHERDIR) $(SHAREDDIR) -name '*.go')

all: $(OUTPUTDIR)/patcher $(OUTPUTDIR)/server $(OUTPUTDIR)/client

$(OUTPUTDIR)/patcher: $(PATCHERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(PATCHERDIR) && \
	go build -o ../$(OUTPUTDIR)/patcher main.go

$(OUTPUTDIR)/server: $(SERVERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(SERVERDIR) && \
	go build -o ../$(OUTPUTDIR)/server main.go

$(OUTPUTDIR)/client: $(CLIENTSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(CLIENTDIR) && \
	go build -o ../$(OUTPUTDIR)/client main.go

$(ASSETDIR)/assets.go: $(ASSETS)
	cd $(CLIENTDIR) && \
	go-bindata -o assets/assets.go -pkg assets -prefix assets/ assets/

.PHONY: clean
	rm -rf bin
	rm $(ASSETDIR)/assets.go