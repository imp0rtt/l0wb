package handlers

import (
	"awesomeProject3/internal/apperror"
	"awesomeProject3/pkg/logging"
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var _ Handler = &handler{}

const (
	URLHome   = "/"
	URLResult = "/result"
)

type handler struct {
	logger *logging.Logger
	cache  *cache.Cache
}

func NewHandler(logger *logging.Logger, cache *cache.Cache) *handler {
	return &handler{logger: logger, cache: cache}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, URLHome, apperror.Middleware(h.handleHome))
	router.HandlerFunc(http.MethodPost, URLResult, apperror.Middleware(h.handleResult))
}

func (h *handler) handleHome(w http.ResponseWriter, r *http.Request) error {
	// Определите путь к файлу HTML
	htmlFilePath := "./public/web/test.html"

	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		http.Error(w, "Ошибка чтения файла HTML", http.StatusInternalServerError)
		return err
	}

	// Установите правильные заголовки Content-Type для HTML и CSS файлов
	w.Header().Set("Content-Type", "text/html")

	// Отправьте HTML содержимое
	w.Write(htmlContent)

	return nil
}

func (h *handler) handleResult(w http.ResponseWriter, r *http.Request) error {
	// Получить параметр "order_uid" из POST-запроса
	r.ParseForm()
	orderUID := r.FormValue("order_uid")

	result, found := h.cache.Get(orderUID)

	// Отправить JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	if !found {
		response := map[string]string{"error": "Not found this order_uid"}
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]interface{}{"result": result}
		json.NewEncoder(w).Encode(response)
	}

	return nil
}
