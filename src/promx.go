package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests")

func main() {

	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infoln(http.ListenAndServe(*addr, nil))
}
