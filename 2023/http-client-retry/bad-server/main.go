package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	serve := func(addr string) {
		fmt.Println("serve on", addr)

		http.DefaultServeMux.HandleFunc("/429", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})
		http.DefaultServeMux.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		http.DefaultServeMux.HandleFunc("/503", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})
		http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("receive %s\n", r.URL.Path)
			time.Sleep(100 * time.Millisecond)
			panic("empty reply")
		})
		http.ListenAndServe(addr, http.DefaultServeMux)
	}

	switch os.Getenv("MODE") {
	case "none":
		select {}
	case "proxy":
		go serve(":8081")
		serverURL, err := url.Parse("http://127.0.0.1:8081")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("proxy :8080 -> :8081")
		pxy := httputil.NewSingleHostReverseProxy(serverURL)
		pxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("http: proxy error: %v\n", err)
			w.WriteHeader(http.StatusBadGateway)
		}
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/504" {
				w.WriteHeader(http.StatusGatewayTimeout)
				return
			}
			if r.URL.Path == "/none" {
				_, err = http.Get("http://127.0.0.1:8082")
				pxy.ErrorHandler(w, r, err)
				return
			}
			pxy.ServeHTTP(w, r)
		}))
	default:
		serve(":8080")
	}
}
