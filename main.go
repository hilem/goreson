package main

import (
	"flag"
  "log"
)

func main() {
	var (
		addr string = "0.0.0.0:3000"
	)

  log.Println("Parsing command arguments")
	flag.StringVar(&addr, "addr", "0.0.0.0:3000", "")
	flag.Parse()

	server := NewServer(addr)
	StartServer(server)
}
