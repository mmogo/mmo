package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fmt"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/pborman/uuid"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var playerID = flag.String("id", "", "player id to use")
var confFile = flag.String("conf", "login.txt", "login config file")

func main() {
	flag.Parse()

	logFile, err := os.Create("game.log")
	if err != nil {
		log.Fatal(err)
	}

	out := io.MultiWriter(logFile, os.Stdout)

	logger := log.New(out, "", log.LstdFlags)

	if *playerID == "" {
		confData, err := ioutil.ReadFile(*confFile)
		if err != nil {
			logger.Fatalf("%s not found: %v", *confFile, err)
		}
		lines := strings.Split(string(confData), "\n")
		for _, line := range lines {
			line = strings.Replace(line, " ", "", -1)
			if strings.Contains(line, "server=") {
				*addr = strings.Replace(line, "server=", "", -1)
			}
			if *playerID == "" && strings.Contains(line, "player_id=") {
				*playerID = strings.Replace(line, "player_id=", "", -1)
			}
		}
		if *playerID == "" {
			*playerID = uuid.New()

			//in case line is blank
			updatedConf := strings.Replace(string(confData), "player_id=", "", -1)

			conf, err := os.Create(*confFile)
			if err != nil {
				logger.Fatal(err)
			}
			if runtime.GOOS == "windows" {
				if _, err := fmt.Fprintf(conf, "%s\r\nplayer_id=%s", updatedConf, *playerID); err != nil {
					logger.Fatal(err)
				}
			} else {
				if _, err := fmt.Fprintf(conf, "%s\nplayer_id=%s", updatedConf, *playerID); err != nil {
					logger.Fatal(err)
				}
			}

		}
	}

	var clientName string
	switch runtime.GOOS {
	case "windows":
		clientName = "client-windows-4.0-amd64.exe"
	case "darwin":
		clientName = "client-darwin-10.6-amd64"
	default:
		clientName = "client-linux-amd64"
	}
	res, err := lxhttpclient.GetAsync(*addr, "/"+clientName, nil)
	if err != nil {
		logger.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	clientBin, err := os.Create(clientName)
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
	cmd := exec.Command(filepath.Join(cwd, clientBin.Name()), "--addr", *addr, "--id", *playerID)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		logger.Fatal(err)
	}
}
