package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/pborman/uuid"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var playerID = flag.String("id", "", "player id to use")
var confFile = flag.String("conf", "login.txt", "login config file")
var protocol = flag.String("protocol", "udp", fmt.Sprintf("network protocol to use."))
var out io.Writer

func main() {
	flag.Parse()
	logFile, err := os.OpenFile("game.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	} else {
		out = io.MultiWriter(logFile, os.Stdout)
		log.SetOutput(out)
	}
	if *playerID == "" {
		*playerID = uuid.New()
	}
	if confData, err := ioutil.ReadFile(*confFile); err == nil {
		log.Printf("reading %q", *confFile)
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
				log.Fatal(err)
			}
			if runtime.GOOS == "windows" {
				if _, err := fmt.Fprintf(conf, "%s\r\nplayer_id=%s", updatedConf, *playerID); err != nil {
					log.Fatal(err)
				}
			} else {
				if _, err := fmt.Fprintf(conf, "%s\nplayer_id=%s", updatedConf, *playerID); err != nil {
					log.Fatal(err)
				}
			}

		}
	}

	gui()

}

func getClientName() string {
	var clientName string
	switch runtime.GOOS {
	case "windows":
		clientName = "client-windows-4.0-amd64.exe"
	case "darwin":
		clientName = "client-darwin-10.6-amd64"
	default:
		clientName = "client-linux-amd64"
	}

	return clientName
}

func runClient(clientName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command(filepath.Join(cwd, clientName), "--addr", *addr, "--id", *playerID, "--protocol", *protocol)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func downloadClient(clientName string) error {
	var checksum string
	if currentClient, err := os.Open(clientName); err == nil {
		defer currentClient.Close()
		h := md5.New()
		if _, err := io.Copy(h, currentClient); err != nil {
			return err
		}
		checksum = fmt.Sprintf("%x", h.Sum(nil))
	}
	query := url.Values{}
	query.Set("checksum", checksum)
	res, err := lxhttpclient.GetAsync(*addr, "/"+clientName+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	//we already have the right client, skip download
	if res.StatusCode == http.StatusNoContent {
		log.Printf("up to date, skipping download")
		return nil
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error while downloading: %s", res.Status)
	}
	log.Printf("downloading client %q", clientName)

	clientBin, err := os.Create(clientName)
	if err != nil {
		return err
	}

	if err := chmod(clientBin, 0755); err != nil {
		return err
	}

	_, err = io.Copy(clientBin, res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := clientBin.Close(); err != nil {
		return err
	}
	log.Println("download complete")
	return nil
}
