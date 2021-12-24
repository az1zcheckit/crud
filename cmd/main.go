package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
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

// creating Server
func execute(host string, port string) (err error) {
	srv := &http.Server{
		Addr: net.JoinHostPort(host, port),
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Print(request.RequestURI) // полный URI
			log.Print(request.Method)
			log.Print(request.Header)
			log.Print(request.Header.Get("Content-Type")) // конкретный заголовок

			log.Print(request.FormValue("tags"))     // только первое значение Query + POST
			log.Print(request.PostFormValue("tags")) // только первое значение POST

			body, err := ioutil.ReadAll(request.Body) // тело запроса
			if err != nil {
				log.Print(err)
			}
			log.Printf("%s", body)

			err = request.ParseMultipartForm(10 * 1024 * 1024) // 10 MB
			if err != nil {
				log.Print(err)
			}

			// доступно только после ParseForm (либо FormValue, PostFormValue)
			log.Print(request.Form)     // все значения формы
			log.Print(request.PostForm) // все значения формы

			log.Print(request.FormFile("image"))
		}),
	}
	// running with ListenAndServe
	log.Print("server is running in "+host, ":"+port+"..")
	return srv.ListenAndServe()
}
