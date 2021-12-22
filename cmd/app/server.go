package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/az1zcheckit/http/pkg/banners"
)

// Server представляет собой логический сервер нашего приложения.
type Server struct {
	mux        *http.ServeMux
	bannersSvc *banners.Service
}

// NewServer - функция-конструктор для создания сервера.
func NewServer(mux *http.ServeMux, bannersSvc *banners.Service) *Server {
	return &Server{mux: mux, bannersSvc: bannersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

// Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	s.mux.HandleFunc("/banners.getAll", s.handleGetAllBanners)
	s.mux.HandleFunc("/banners.getById", s.handleGetPostByID)
	s.mux.HandleFunc("/banners.save", s.handleSaveBanner)
	s.mux.HandleFunc("/banners.removeById", s.handleremoveByID)
}

// handleGetAllBanners - ...
func (s *Server) handleGetAllBanners(writer http.ResponseWriter, request *http.Request) {
	all, err := s.bannersSvc.All(request.Context())
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
	log.Print("ready")
	_, err = writer.Write([]byte(data))
	if err != nil {
		log.Print("Error!: Can't write anything on data.")
	}
}

// handleGetPostByID - по id мы хотим отдавать баннер в формате JSON:
func (s *Server) handleGetPostByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.bannersSvc.ByID(request.Context(), id)
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

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

// handleSaveBanner - сохраняет или обновляет баннер.
func (s *Server) handleSaveBanner(writer http.ResponseWriter, request *http.Request) {
	// добавление всех параметров баннера (C.R.U.D)
	idParam := request.URL.Query().Get("id")
	titleParam := request.URL.Query().Get("title")
	contentParam := request.URL.Query().Get("content")
	buttonParam := request.URL.Query().Get("button")
	linkParam := request.URL.Query().Get("link")

	ID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	Title := string(titleParam)
	Content := string(contentParam)
	Button := string(buttonParam)
	Link := string(linkParam)
	// Сохранение/Обновление баннера
	banner := banners.Banner{
		ID:      ID,
		Title:   Title,
		Content: Content,
		Button:  Button,
		Link:    Link,
	}
	bannerRes, err := s.bannersSvc.Save(request.Context(), &banner)
	if err != nil {
		log.Print()
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(bannerRes)
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
}

// handleremoveByID - удаляет баннер по идентификатору.
func (s *Server) handleremoveByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	delBanner, err := s.bannersSvc.RemoveByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(delBanner)
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
}
