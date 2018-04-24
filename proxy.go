package main

import (
	"net/http"
	"net/http/httputil"

	"github.com/julienschmidt/httprouter"
)

// ReverseProxy reverse proxy handler
func ReverseProxy(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	reverseProxy := httputil.NewSingleHostReverseProxy(globalConfig.remote)

	reverseProxy.ServeHTTP(w, r)
}
