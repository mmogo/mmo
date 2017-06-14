package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hectane/go-acl"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"runtime"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	//TODO: separtae by platform / architecture (in request)
	clientName := "client"
	if runtime.GOOS == "windows" {
		clientName = "client.exe"
	}
	res, err := lxhttpclient.GetAsync(*addr, "/client/"+clientName, nil)
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
	if runtime.GOOS == "windows" {
		if err := acl.Chmod(clientBin.Name(), 0755); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := clientBin.Chmod(0755); err != nil {
			log.Fatal(err)
		}
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
