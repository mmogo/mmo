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

all: $(OUTPUTDIR)/patcher \
	 $(OUTPUTDIR)/server \
	 $(OUTPUTDIR)/client \
	 $(OUTPUTDIR)/client-windows-4.0-amd64.exe \
	 $(OUTPUTDIR)/patcher-windows-4.0-amd64.exe

$(OUTPUTDIR)/patcher: $(PATCHERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(PATCHERDIR) && \
	go build -o ../$(OUTPUTDIR)/patcher .

$(OUTPUTDIR)/server: $(SERVERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(SERVERDIR) && \
	go build -o ../$(OUTPUTDIR)/server .

$(OUTPUTDIR)/client: $(CLIENTSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(CLIENTDIR) && \
	go build -o ../$(OUTPUTDIR)/client .

$(ASSETDIR)/assets.go: $(ASSETS)
	cd $(CLIENTDIR) && \
	go-bindata -o assets/assets.go -pkg assets -prefix assets/ assets/

$(OUTPUTDIR)/client-windows-4.0-amd64.exe: $(CLIENTSOURCES)
	xgo -dest=bin -targets=windows/amd64 -pkg ./client .

$(OUTPUTDIR)/patcher-windows-4.0-amd64.exe: $(PATCHERSOURCES)
	xgo -dest=bin -targets=windows/amd64 -pkg ./patcher .

.PHONY: clean
	rm -rf bin
	rm $(ASSETDIR)/assets.go