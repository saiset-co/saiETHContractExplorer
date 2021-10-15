package main

import (
	"net/http"
	"strings"
)

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func api(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	err := r.ParseForm()

	if err != nil {
		return
	}

	method := strings.Join(r.Form["method"], "")
	switch method {
	default:
		{
			_, _ = w.Write([]byte("I'm alive"))
		}
	}
}
