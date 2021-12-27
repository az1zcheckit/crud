package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/az1zcheckit/crud/pkg/customers"
)

// Server представляет собой логический сервер нашего приложения.
type Server struct {
	mux          *http.ServeMux
	customersSvc *customers.Service
}

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *http.ServeMux, customersSvc *customers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

// Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	s.mux.HandleFunc("/customers.getAll", s.handleGetAllCustomers)
	s.mux.HandleFunc("/customers.getAllActive", s.handleGetAllActiveCustomers)
	s.mux.HandleFunc("/customers.getById", s.handleGetCustomerByID)
	s.mux.HandleFunc("/customers.save", s.handleSaveCustomers)
	s.mux.HandleFunc("/customers.removeById", s.handleRemoveByID)
	s.mux.HandleFunc("/customers.blockById", s.handleBlockByID)
	s.mux.HandleFunc("/customers.unblockById", s.handleUnBlockByID)
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
}

// handleGetCustomerByID - нахождение покупателя по id.
func (s *Server) handleGetCustomerByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")

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
}

// handleSaveBanner - создаёт или обновляет покупателей .
func (s *Server) handleSaveCustomers(writer http.ResponseWriter, request *http.Request) {
	// добавление всех параметров баннера (C.R.U.D)
	idParam := request.FormValue("id")
	Name := request.FormValue("name")
	Phone := request.FormValue("phone")
	Active := request.FormValue("active")

	ID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	act := true
	log.Print(Active)
	if Active != "true" {
		act = false
	}

	customer := customers.Customer{
		ID:      ID,
		Name:    Name,
		Phone:   Phone,
		Active:  act,
		Created: time.Now(),
	}
	customersRes, err := s.customersSvc.Save(request.Context(), &customer)
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
	return
}

// handleremoveByID - удаляет покупателя по идентификатору.
func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")

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
	idParam := request.URL.Query().Get("id")
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
}

// handleUnBlockById - выставляет статус active в true.
func (s *Server) handleUnBlockByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")
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
}
