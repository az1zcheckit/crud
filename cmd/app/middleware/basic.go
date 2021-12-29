package middleware

import (
	"log"
	"net/http"
)

// Basic - middleware, basic это функция для авторизации.
func Basic(auth func(login, pass string) bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			username, password, ok := request.BasicAuth()
			if !ok {
				log.Print("Can't parse username and password")
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if !auth(username, password) {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			handler.ServeHTTP(writer, request)
		})
	}
}
