package main

import (
	_ "image/png"

	"flag"
	"fmt"
	"log"

	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/pkg/client/api"
	"github.com/mmogo/mmo/pkg/client/debug"
	"github.com/mmogo/mmo/pkg/client/full"
	"github.com/mmogo/mmo/pkg/shared"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {
	addr := flag.String("addr", "localhost:8080", "address of server")
	id := flag.String("id", "", "playerid to use")
	protocol := flag.String("protocol", "udp", fmt.Sprintf("network protocol to use. available %s | %s", shared.ProtocolTCP, shared.ProtocolUDP))
	debugMode := flag.Bool("d", true, "run debug version of client")
	flag.Parse()
	if *id == "" {
		log.Fatal("id must be provided")
	}

	f, err := os.Create("cpuprofile")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			pprof.StopCPUProfile()
			log.Fatalf("detected sig: %s, shutting down", sig)
		}
	}()
	pixelgl.Run(func() {
		if err := run(*protocol, *addr, *id, *debugMode); err != nil {
			log.Fatal(err)
		}
	})
}

func run(protocol, addr, id string, debugMode bool) error {
	conn, world, err := api.Dial(protocol, addr, id)
	if err != nil {
		return errors.New("failed to dial server", err)
	}

	//start client
	if debugMode {
		debug.NewClient(id, conn, world).Run()
	} else {
		client := full.NewClient(id, conn, world)
		if err := client.Init(); err != nil {
			return errors.New("failed to initialize full client", err)
		}
		client.Run()
	}

	return errors.New("client exited for unknown reason", nil)
}
