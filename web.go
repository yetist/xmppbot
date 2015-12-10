package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type WebServer struct {
	Dispatcher *mux.Router
	Urls       map[string]func(w http.ResponseWriter, r *http.Request)
	Port       string
}

func NewWebServer(port int) *WebServer {
	return &WebServer{
		Port:       strconv.Itoa(port),
		Dispatcher: mux.NewRouter(),
		Urls:       make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
}

func (s *WebServer) Handler(path string, handler http.HandlerFunc, name string) {
	s.Urls[name] = handler
	s.Dispatcher.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if s.Urls[name] != nil {
			s.Urls[name](w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

func (s *WebServer) Destroy(name string) {
	s.Urls[name] = nil
}

func (s *WebServer) Start() {
	http.ListenAndServe(":"+s.Port, s.Dispatcher)
}
