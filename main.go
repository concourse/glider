package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/winston-ci/redgreen/api"
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:5637",
	"listening address",
)

var peerAddr = flag.String(
	"peerAddr",
	"",
	"address reachable by other nodes",
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

	logger := log.New(os.Stdout, "", 0)

	handler := api.New(logger, *peerAddr, *proleURL)

	err := http.ListenAndServe(*listenAddr, handler)
	logger.Fatalln("listen error:", err)
}
