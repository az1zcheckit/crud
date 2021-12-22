package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/az1zcheckit/http/cmd/app"
	"github.com/az1zcheckit/http/pkg/banners"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

// creating Server
func execute(host string, port string) (err error) {
	mux := http.NewServeMux()
	bannersSvc := banners.NewService()
	server := app.NewServer(mux, bannersSvc)
	server.Init()

	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: server,
	}
	// running with ListenAndServe
	log.Print("server is running in "+host, ":"+port+"..")
	return srv.ListenAndServe()
}

type handler struct {
	mu       *sync.RWMutex
	handlers map[string]http.HandlerFunc
}

// ServeHTTP обрабатывает все запросы
func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.mu.RLock()
	handler, ok := h.handlers[request.URL.Path]
	h.mu.RUnlock()
	if !ok {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	handler(writer, request)
}
