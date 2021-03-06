package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/az1zcheckit/crud/cmd/app"
	"github.com/az1zcheckit/crud/pkg/customers"
	"github.com/az1zcheckit/crud/pkg/security"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	// адрес подключения
	//протокол://логин:палоь@хост:порт/бд
	dsn := "postges://app:pass@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	// 	password := "secret"
	// 	// hash := md5.New()
	// 	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// 	if err != nil {
	// 		log.Print(err)
	// 		os.Exit(1)
	// 	}
	// 	log.Print(hex.EncodeToString(hash))

	// 	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	// 	if err != nil {
	// 		log.Print("password is invalid")
	// 		os.Exit(1)
	// 	}
	// 	// salted := append(salt, []byte(password)...)
	// 	// hash := md5.New()
	// 	// _, err = hash.Write(salted)
	// 	// if err != nil {
	// 	// 	log.Print(err)
	// 	// 	os.Exit(1)
	// 	// }
	// 	// // считаем хэш
	// 	// sum := hash.Sum([]byte(nil))
	// 	// log.Print(hex.EncodeToString(sum)) // можно просто log.Printf("%x", sum)
}

func execute(host string, port string, dsn string) (err error) {
	// создание контейнера где будем хранить все методы и функции.
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter, // mux -> "github.com/gorilla/mux"
		func() (*pgxpool.Pool, error) {
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return pgxpool.Connect(ctx, dsn)
		},
		customers.NewService,
		security.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}
	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		log.Print(err)
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		log.Print("server is running in " + host + ":" + port + "..")
		return server.ListenAndServe()
	})
}
