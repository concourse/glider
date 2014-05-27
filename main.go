package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/winston-ci/redgreen/api"
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:5637",
	"listening address",
)

var peerAddr = flag.String(
	"peerAddr",
	"127.0.0.1:5637",
	"external address of the redgreen server, used for callbacks",
)

var proleURL = flag.String(
	"proleURL",
	"http://127.0.0.1:4637",
	"address denoting the prole service",
)

func main() {
	flag.Parse()

	if *peerAddr == "" {
		log.Fatalln("must specify -peerAddr")
	}

	handler, err := api.New(*peerAddr, *proleURL)
	if err != nil {
		log.Fatalln("failed to initialize handler:", err)
	}

	err = http.ListenAndServe(*listenAddr, handler)
	log.Fatalln("listen error:", err)
}
