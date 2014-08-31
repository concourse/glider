package main

import (
	"flag"
	"os"

	"github.com/concourse/glider/api"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:5637",
	"listening address",
)

var peerAddr = flag.String(
	"peerAddr",
	"127.0.0.1:5637",
	"external address of the glider server, used for callbacks",
)

var turbineURL = flag.String(
	"turbineURL",
	"http://127.0.0.1:4637",
	"address denoting the turbine service",
)

func main() {
	flag.Parse()

	logger := lager.NewLogger("glider")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	handler, err := api.New(logger.Session("api"), *peerAddr, *turbineURL)
	if err != nil {
		logger.Fatal("failed-to-initialize-handler", err)
	}

	server := http_server.New(*listenAddr, handler)
	running := ifrit.Envoke(sigmon.New(server))

	logger.Info("listening", lager.Data{
		"api": *listenAddr,
	})

	err = <-running.Wait()
	if err == nil {
		logger.Info("exited")
	} else {
		logger.Error("failed", err)
		os.Exit(1)
	}
}
