package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	_ "net/http/pprof"
)

func pprofEndpoints(router *httprouter.Router) error {
	router.Handler(http.MethodGet, "/debug/pprof/*item", http.DefaultServeMux)
	return nil
}

func prometheusMetrics(router *httprouter.Router) error {
	router.Handler(http.MethodGet,"/metrics", promhttp.Handler())
	return nil
}