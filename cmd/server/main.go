package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mmogo/mmo/pkg/server"
	"github.com/mmogo/mmo/pkg/shared"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {
	port := flag.Int("port", 8080, "port to serve on")
	protocol := flag.String("protocol", "udp", fmt.Sprintf("network protocol to use. available %s | %s", shared.ProtocolTCP, shared.ProtocolUDP))
	flag.Parse()
	errc := make(chan error)
	go func() { log.Fatal(server.NewMMOServer().Run(*protocol, *port, errc)) }()
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
