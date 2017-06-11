package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/layer-x/layerx-commons/lxhttpclient"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	//TODO: separtae by platform / architecture (in request)
	res, err := lxhttpclient.GetAsync(*addr, "/client/client", nil)
	if err != nil {
		log.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	clientBin, err := os.Create("client.bin")
	if err != nil {
		log.Fatal(err)
	}
	if err := clientBin.Chmod(0755); err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(clientBin, res.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if err := clientBin.Close(); err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(cwd, clientBin.Name()), "--addr", *addr)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
