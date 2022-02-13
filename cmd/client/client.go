package main

import (
	"cjvirtucio87/tftp-go/pkg/tftp"
	"flag"
	"log"
)

var (
	address       = flag.String("address", "127.0.0.1:69", "server address")
	clientAddress = flag.String("client-address", "127.0.0.2:69", "client address")
	filename      = flag.String("filename", "", "filename of the requested payload")
)

func main() {
	flag.Parse()

	if *filename == "" {
		log.Fatal("filename must not be empty")
	}

	c := tftp.Client{}
	log.Fatal(c.Send(*clientAddress, *address, *filename))
}