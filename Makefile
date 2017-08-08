SOURCEDIR=$(shell pwd)

CMDDIR=$(SOURCEDIR)/cmd
CLIENTDIR=$(CMDDIR)/client
SERVERDIR=$(CMDDIR)/server
PATCHERDIR=$(CMDDIR)/patcher

PKGDIR=$(SOURCEDIR)/pkg
SHAREDDIR=$(PKGDIR)/shared

ASSETDIR=$(SOURCEDIR)/assets
ASSETS := $(shell find $(ASSETDIR))
ASSETFILE := $(PKGDIR)/assets/assets.go

OUTPUTDIR := $(SOURCEDIR)/bin

CLIENTSOURCES := $(shell find $(CLIENTDIR) $(SHAREDDIR) $(PKGDIR)/client -name '*.go') $(ASSETFILE)
SERVERSOURCES := $(shell find $(SERVERDIR) $(SHAREDDIR) $(PKGDIR)/server -name '*.go')
PATCHERSOURCES := $(shell find $(PATCHERDIR) -name '*.go')

IMAGE=ilackarms/xgo-latest

default: linux

all: linux windows darwin

linux: $(OUTPUTDIR)/patcher-linux-amd64 \
       $(OUTPUTDIR)/server-linux-amd64 \
       $(OUTPUTDIR)/client-linux-amd64 \
       $(ASSETFILE)

windows: $(OUTPUTDIR)/client-windows-4.0-amd64.exe \
         $(OUTPUTDIR)/patcher-windows-4.0-amd64.exe	\
         $(ASSETFILE)

darwin: $(OUTPUTDIR)/client-darwin-10.6-amd64 \
	    $(OUTPUTDIR)/patcher-darwin-10.6-amd64 \
		$(ASSETFILE)

$(OUTPUTDIR)/patcher-linux-amd64: $(PATCHERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(PATCHERDIR) && \
	go build -o $@ .

$(OUTPUTDIR)/server-linux-amd64: $(SERVERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(SERVERDIR) && \
	go build -o $@ .

$(OUTPUTDIR)/client-linux-amd64: $(CLIENTSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(CLIENTDIR) && \
	go build -o $@ .

$(ASSETFILE) : $(ASSETS)
	go-bindata -o $(ASSETFILE) -pkg assets -prefix $(ASSETDIR) $(ASSETDIR)/...

$(OUTPUTDIR)/client-windows-4.0-amd64.exe: $(CLIENTSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=windows/amd64 -pkg ./client .

$(OUTPUTDIR)/patcher-windows-4.0-amd64.exe: $(PATCHERSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=windows/amd64 -pkg ./patcher .

$(OUTPUTDIR)/client-darwin-10.6-amd64: $(CLIENTSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=darwin/amd64 -pkg ./client .

$(OUTPUTDIR)/patcher-darwin-10.6-amd64: $(PATCHERSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=darwin/amd64 -pkg ./patcher .

.PHONY: clean

clean:
	rm -rf bin
