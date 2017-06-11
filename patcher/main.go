package main

import (
	"flag"
	"log"
	"os"

	"github.com/layer-x/layerx-commons/lxhttpclient"
	"io"
	"plugin"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	//TODO: separtae by platform / architecture (in request)
	res, err := lxhttpclient.GetAsync(*addr, "/client/client.so", nil)
	if err != nil {
		log.Fatal(err)
	}
	clientSO, err := os.Create("client.so")
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(clientSO, res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	client, err := plugin.Open(clientSO.Name())
	if err != nil {
		log.Fatal(err)
	}
	clientMain, err := client.Lookup("Main")
	if err != nil {
		log.Fatal(err)
	}
	clientMain.(func(addr string))(*addr)
}
