SOURCEDIR=.
SHAREDDIR=$(SOURCEDIR)/shared
CLIENTDIR=$(SOURCEDIR)/client
SERVERDIR=$(SOURCEDIR)/server
PATCHERDIR=$(SOURCEDIR)/patcher
ASSETDIR=$(CLIENTDIR)/assets
ASSETS := $(shell find $(SOURCEDIR)/client/assets -name assets.go -prune -o -print)
OUTPUTDIR := $(SOURCEDIR)/bin

SERVERADDR := localhost:8080

CLIENTSOURCES := $(shell find $(CLIENTDIR) $(SHAREDDIR) -name '*.go') $(ASSETDIR)/assets.go
SERVERSOURCES := $(shell find $(SERVERDIR) $(SHAREDDIR) -name '*.go')
PATCHERSOURCES := $(shell find $(PATCHERDIR) -name '*.go')

IMAGE=ilackarms/xgo-latest

default: $(ASSETDIR)/assets.go linux

all: $(ASSETDIR)/assets.go linux windows darwin

linux: $(OUTPUTDIR)/patcher-linux-amd64 \
       $(OUTPUTDIR)/server-linux-amd64 \
       $(OUTPUTDIR)/client-linux-amd64 \
       $(OUTPUTDIR)/login.txt

windows: $(OUTPUTDIR)/client-windows-4.0-amd64.exe \
         $(OUTPUTDIR)/patcher-windows-4.0-amd64.exe	\
         $(OUTPUTDIR)/login.txt

darwin: $(OUTPUTDIR)/client-darwin-10.6-amd64 \
	    $(OUTPUTDIR)/patcher-darwin-10.6-amd64 \
		$(OUTPUTDIR)/login.txt

$(OUTPUTDIR)/patcher-linux-amd64: $(PATCHERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(PATCHERDIR) && \
	go build -o ../$@ .

$(OUTPUTDIR)/server-linux-amd64: $(SERVERSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(SERVERDIR) && \
	go build -o ../$@ .

$(OUTPUTDIR)/client-linux-amd64: $(CLIENTSOURCES)
	mkdir -p $(OUTPUTDIR)
	cd $(CLIENTDIR) && \
	go build -o ../$@ .

$(ASSETDIR)/assets.go: $(ASSETS)
	cd $(CLIENTDIR) && \
	go-bindata -o assets/assets.go -pkg assets -prefix assets/ assets/...

$(OUTPUTDIR)/client-windows-4.0-amd64.exe: $(CLIENTSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=windows/amd64 -pkg ./client .

$(OUTPUTDIR)/patcher-windows-4.0-amd64.exe: $(PATCHERSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=windows/amd64 -pkg ./patcher .

$(OUTPUTDIR)/client-darwin-10.6-amd64: $(CLIENTSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=darwin/amd64 -pkg ./client .

$(OUTPUTDIR)/patcher-darwin-10.6-amd64: $(PATCHERSOURCES)
	xgo -image $(IMAGE) -dest=bin -targets=darwin/amd64 -pkg ./patcher .

$(OUTPUTDIR)/login.txt:
	echo "server=$(SERVERADDR)" > $@

.PHONY: clean

clean:
	rm -rf bin
