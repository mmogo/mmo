// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package main

import (
	_ "image/png"

	"flag"
	"fmt"
	"image/color"
	"log"
	"net"

	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/shared"
	"github.com/xtaci/smux"
	"golang.org/x/image/colornames"
)

const (
	UP        = shared.UP
	DOWN      = shared.DOWN
	LEFT      = shared.LEFT
	RIGHT     = shared.RIGHT
	UPLEFT    = shared.UPLEFT
	UPRIGHT   = shared.UPRIGHT
	DOWNLEFT  = shared.DOWNLEFT
	DOWNRIGHT = shared.DOWNRIGHT
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {
	addr := flag.String("addr", "localhost:8080", "address of server")
	id := flag.String("id", "", "playerid to use")
	protocol := flag.String("protocol", "udp", fmt.Sprintf("network protocol to use. available %s | %s", shared.ProtocolTCP, shared.ProtocolUDP))
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
		if err := run(*protocol, *addr, *id); err != nil {
			log.Fatal(err)
		}
	})
}

func run(protocol, addr, id string) error {
	conn, err := dialServer(protocol, addr, id)
	if err != nil {
		return errors.New("failed to dial server", err)
	}

	// start window
	cfg := pixelgl.WindowConfig{
		Title:  "loading",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return fmt.Errorf("creating window: %v", err)
	}

	// sync with server
	msg, err := shared.GetMessage(conn)
	if err != nil {
		return errors.New("failed reading message", err)
	}

	if msg.Update == nil || msg.Update.WorldState == nil || msg.Update.WorldState.World == nil {
		return errors.New("expected Sync message on server handshake, got "+msg.String(), nil)
	}

	//start client
	newClient(id, conn, win, msg.Update.WorldState.World).start()

	return errors.New("client exited for unknown reason", nil)
}

func dialServer(protocol, addr, id string) (net.Conn, error) {
	log.Printf("dialing %s", addr)
	conn, err := shared.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}
	session, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return nil, err
	}
	stream, err := session.OpenStream()
	if err != nil {
		return nil, err
	}
	conn = stream

	if err := shared.SendMessage(&shared.Message{
		Request: &shared.Request{
			ConnectRequest: &shared.ConnectRequest{
				ID: id,
			},
		}}, conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func stringToColor(str string) color.Color {
	colornum := 0
	for _, s := range str {
		colornum += int(s)
	}
	all := len(colornames.Names)
	name := colornames.Names[colornum%all]
	return colornames.Map[name]
}
