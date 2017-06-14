package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/layer-x/layerx-commons/lxhttpclient"
	"io/ioutil"
	"runtime"
	"strings"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var conffile = flag.String("conf", "login.txt", "login config file")

func main() {
	flag.Parse()

	//override addr with login.txt
	confdata, err := ioutil.ReadFile(*conffile)
	if err == nil {
		lines := strings.Split(string(confdata), "\n")
		for _, line := range lines {
			line = strings.Replace(line, " ", "", -1)
			if strings.Contains(line, "server=") {
				*addr = strings.Replace(line, "server=", "", -1)
				break
			}
		}
	}

	//TODO: separtae by platform / architecture (in request)
	clientName := "client"
	if runtime.GOOS == "windows" {
		clientName = "client-windows-4.0-amd64.exe"
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

	if err := chmod(clientBin, 0755); err != nil {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
