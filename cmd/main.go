package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/az1zcheckit/crud/cmd/app"
	"github.com/az1zcheckit/crud/pkg/customers"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	dsn := "postgres://app:postgres@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

type handler struct {
	mu       *sync.RWMutex
	handlers map[string]http.HandlerFunc
}

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

func execute(host string, port string, dsn string) (err error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()
	// TODO запросы
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS customers (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		log.Print(err)
		return
	}

	mux := http.NewServeMux()
	customersSvc := customers.NewService(db)

	server := app.NewServer(mux, customersSvc)
	server.Init()

	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: server,
	}

	log.Print("server start " + host + ":" + port)
	return srv.ListenAndServe()
}
