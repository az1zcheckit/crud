package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/az1zcheckit/crud/pkg/customers"
	"github.com/az1zcheckit/crud/pkg/security"
	"github.com/gorilla/mux"
)

var ErrNotFound = errors.New("item not found")

// ErrNoSuchUser если пользоваетль не найден.
var ErrNoSuchUser = errors.New("No such user")

// ErrInvalidPassword если пароль не верный
var ErrInvalidPassword = errors.New("Invalid password")

// ErrInternal если происходить другая ошибка
var ErrInternal = errors.New("Internal error")

// ErrExpired исчерпал свой токен
var ErrExpired = errors.New("Token is expired")

// Server представляет собой логический сервер нашего приложения.
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	securitySvc  *security.Service
}

// Token..
type Token struct {
	Token string `json:"token"`
}

// Responce..
type Responce struct {
	CustomerID int64  `json:"customerId"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

// ResponceOk..
type ResponceOk struct {
	Status     string `json:"status"`
	CustomerID int64  `json:"customerId"`
}

// ResponceFail..
type ResponceFail struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *mux.Router, customersSvc *customers.Service, securitySvc *security.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, securitySvc: securitySvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

// save done
// active done
// block done
// unblock done

// Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	//s.mux.HandleFunc("/customers.getAll", s.handleGetAllCustomers)
	s.mux.HandleFunc("/customers", s.handleGetAllCustomers).Methods(GET)
	//s.mux.HandleFunc("/customers.getAllActive", s.handleGetAllActiveCustomers)
	s.mux.HandleFunc("/customers/active", s.handleGetAllActiveCustomers).Methods(GET)
	///s.mux.HandleFunc("/customers.getById", s.handleGetCustomerByID)
	s.mux.HandleFunc("/customers/{id}", s.handleGetCustomersByID).Methods(GET)
	//s.mux.HandleFunc("/customers.save", s.handleSaveCustomers)
	s.mux.HandleFunc("/customers", s.handleSaveCustomers).Methods(POST)
	//s.mux.HandleFunc("/customers.removeById", s.handleRemoveByID)
	s.mux.HandleFunc("/customers/{id}", s.handleRemoveByID).Methods(DELETE)
	//s.mux.HandleFunc("/customers.blockById", s.handleBlockByID)
	s.mux.HandleFunc("/customers/{id}/block", s.handleBlockByID).Methods(POST)
	//s.mux.HandleFunc("/customers.unblockById", s.handleUnBlockByID)
	s.mux.HandleFunc("/customers/{id}/block", s.handleUnBlockByID).Methods(DELETE)
	// s.mux.Use(middleware.Basic(s.securitySvc.Auth))

	s.mux.HandleFunc("/api/customers", s.SaveCustomers).Methods(POST)
	s.mux.HandleFunc("/api/customers/token", s.handleGetToken).Methods(POST)
	s.mux.HandleFunc("/api/customers/token/validate", s.handleValidateToken).Methods(POST)
}

func (s *Server) handleGetToken(writer http.ResponseWriter, request *http.Request) {
	var auth *security.Auth
	var tok Token
	err := json.NewDecoder(request.Body).Decode(&auth)
	log.Print(auth)
	if err != nil {
		log.Print("Can't decode login and password")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("Login: ", auth.Login, "Password: ", auth.Password)

	token, err := s.customersSvc.TokenForCustomer(request.Context(), auth.Login, auth.Password)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tok.Token = token
	data, err := json.Marshal(tok)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleValidateToken(writer http.ResponseWriter, request *http.Request) {
	var fail ResponceFail
	var ok ResponceOk
	var token Token
	var data []byte
	code := 200

	err := json.NewDecoder(request.Body).Decode(&token)
	if err != nil {
		log.Print("Can't decode token")
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, er := s.securitySvc.AuthForCustomer(request.Context(), token.Token)

	if er == security.ErrNoSuchUser {
		code = 404
		fail.Status = "fail"
		fail.Reason = "not found"
	} else if er == security.ErrExpiredToken {
		code = 400
		fail.Status = "fail"
		fail.Reason = "expired"
	} else if er == nil {
		log.Print(id)
		ok.Status = "ok"
		ok.CustomerID = id
	} else {
		log.Print("err", er)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	if code != 200 {
		writer.WriteHeader(code)

		data, err = json.Marshal(fail)
		if err != nil {
			log.Print(err)
		}
	} else {
		data, err = json.Marshal(ok)
		if err != nil {
			log.Print(err)
		}
	}
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	return
}

func (s *Server) SaveCustomers(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	err := json.NewDecoder(request.Body).Decode(&item)
	log.Print(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	customer, err := s.customersSvc.SaveCustomer(request.Context(), item)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(customer)

	data, err := json.Marshal(customer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
	http.Error(writer, http.StatusText(http.StatusOK), http.StatusOK)
	return
}

// handleGetAllCustomers берет всю инфу о покупателе..
func (s *Server) handleGetAllCustomers(writer http.ResponseWriter, request *http.Request) {
	all, err := s.customersSvc.All(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	data, err := json.Marshal(all)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("server's working perfect")
	_, err = writer.Write([]byte(data))
	if err != nil {
		log.Print("Error!: Can't write anything on data.")
	}
}

// handleGetAllActiveCustomers - вся инфа об активных покупателей.
func (s *Server) handleGetAllActiveCustomers(writer http.ResponseWriter, request *http.Request) {
	allActive, err := s.customersSvc.AllActive(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	data, err := json.Marshal(allActive)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("server's working perfect")
	_, err = writer.Write([]byte(data))
	if err != nil {
		log.Print("Error!: Can't write anything on data.")
	}
	// //var items []*customers.Customer
	// // чтение данных из файла json
	// items, err := s.customersSvc.AllActive(request.Context())
	// err = json.NewDecoder(request.Body).Decode(&items)
	// if err != nil {
	// 	log.Print(err)
	// 	http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// 	return
	// }
}

/*// handleGetAllActiveCustomer - получает данные всех активных пользователей.
func (s *Server) handleGetAllActiveCustomer(writer http.ResponseWriter, request *http.Request) {
	var items []*customers.Customer
	// чтение данных из файла json
	err := json.NewDecoder(request.Body).Decode(&items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	items, err = s.customersSvc.AllActive(request.Context())
}*/

// handleGetCustomerByID - нахождение покупателя по id.
func (s *Server) handleGetCustomersByID(writer http.ResponseWriter, request *http.Request) {
	//idParam := request.URL.Query().Get("id")
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.customersSvc.ByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("server's working perfect")
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}

	// чтение данных из файла json
	/*item, err = s.customersSvc.ByID(request.Context(), item.ID)
	err = json.NewDecoder(request.Body).Decode(&item.ID)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}*/

	//item, err = s.customersSvc.ByID(request.Context(), item.ID)
}

/*// handleGetAllActiveCustomer - получает данные всех активных пользователей.
func (s *Server) handleGetCustomersByID(writer http.ResponseWriter, request *http.Request) {
	var items []*customers.Customer
	// чтение данных из файла json
	err := json.NewDecoder(request.Body).Decode(&items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	items, err = s.customersSvc.AllActive(request.Context())
}*/

// handleSaveBanner - создаёт или обновляет покупателей .
func (s *Server) handleSaveCustomers(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	customersRes, err := s.customersSvc.Save(request.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(customersRes)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("server's working perfect")
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
	//item, err = s.customersSvc.Save(request.Context(), item)

}

/*// handleSaveCustomer - сохраняет/обновляет данные клиента
func (s *Server) handleSaveCustomer(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Customer
	// чтение данных из файла json
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err = s.customersSvc.Save(request.Context(), item)
}*/

// handleremoveByID - удаляет покупателя по идентификатору.
func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]

	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.RemoveByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	/*	data, err := json.Marshal(rem)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write(data)
		if err != nil {
			log.Print(err)
		}*/
}

// handleBlockById - выставляет статус active в false.
func (s *Server) handleBlockByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.BlockByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// err = s.customersSvc.BlockByID(request.Context(), item.ID)
}

// handleUnBlockById - выставляет статус active в true.
func (s *Server) handleUnBlockByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.customersSvc.UnBlockByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// err = s.customersSvc.UnBlockByID(request.Context(), item.ID)
}
