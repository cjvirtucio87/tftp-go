package main

import (
	"cjvirtucio87/tftp-go/pkg/tftp"
	"flag"
	"io/ioutil"
	"log"
)

var (
	address  = flag.String("address", "127.0.0.1:69", "listen address")
	filepath = flag.String("filepath", "", "filepath to the payload")
)

func main() {
	flag.Parse()

	if *filepath == "" {
		log.Fatal("filepath must not be empty")
	}

	p, err := ioutil.ReadFile(*filepath)
	if err != nil {
		log.Fatal(err)
	}

	s := tftp.Server{
        Logger: tftp.NewZapLogger(),
		Payload: p,
	}

	err = s.ListenAndServe(*address)
	if err != nil {
		log.Fatal(err)
	}
}
