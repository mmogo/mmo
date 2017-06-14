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

	logFile, err := os.Create("game.log")
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

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
		logger.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	clientBin, err := os.Create("client.bin")
	if err != nil {
		logger.Fatal(err)
	}

	if err := chmod(clientBin, 0755); err != nil {
		logger.Fatal(err)
	}

	_, err = io.Copy(clientBin, res.Body)
	if err != nil {
		logger.Fatal(err)
	}
	defer res.Body.Close()
	if err := clientBin.Close(); err != nil {
		logger.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(cwd, clientBin.Name()), "--addr", *addr)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Run(); err != nil {
		logger.Fatal(err)
	}
}
