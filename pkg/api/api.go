package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"GoNews/pkg/db"

	"github.com/gorilla/mux"
)

// API представляет API сервер
type API struct {
	r  *mux.Router
	db *db.Storage
}

// New создает новый API сервер
func New(db *db.Storage) *API {
	api := API{
		r:  mux.NewRouter(),
		db: db,
	}
	api.endpoints()
	return &api
}

// Router возвращает маршрутизатор
func (api *API) Router() *mux.Router {
	return api.r
}

// Регистрация обработчиков
func (api *API) endpoints() {
	api.r.Use(api.corsMiddleware)
	api.r.HandleFunc("/news/{n}", api.posts).Methods(http.MethodGet)
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// Middleware для CORS
func (api *API) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// posts возвращает список новостей
func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(mux.Vars(r)["n"])
	if err != nil {
		http.Error(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	if n <= 0 {
		http.Error(w, "n must be positive", http.StatusBadRequest)
		return
	}

	posts, err := api.db.GetPosts(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
	}
}
