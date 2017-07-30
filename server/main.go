// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mmogo/mmo/shared"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

const (
	ticksPerSecond = 10
	tickTime       = 1.0 / ticksPerSecond

	messagePerTickLimit = 60
)

func main() {
	port := flag.Int("port", 8080, "port to serve on")
	protocol := flag.String("protocol", "udp", fmt.Sprintf("network protocol to use. available %s | %s", shared.ProtocolTCP, shared.ProtocolUDP))
	flag.Parse()
	errc := make(chan error)
	server := newMMOServer()
	go func() { log.Fatal(server.start(*protocol, *port, errc)) }()
	for {
		select {
		case err := <-errc:
			if shared.IsFatal(err) {
				log.Fatal(err)
			}
			log.Println("error:", err)
		}
	}
}
