package main

import (
	"flag"
	"log"
	"os"

	"github.com/rcrowley/go-tigertonic"
	"github.com/winston-ci/redgreen/api"
)

var listenAddr = flag.String(
	"listen",
	"0.0.0.0:5637",
	"listening address",
)

var proleURL = flag.String(
	"proleAddr",
	"http://127.0.0.1:4637",
	"address denoting the prole service",
)

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "", 0)

	handler := api.New(logger, *proleURL)

	server := tigertonic.NewServer(*listenAddr, handler)

	err := server.ListenAndServe()
	logger.Fatalln("listen error:", err)
}
