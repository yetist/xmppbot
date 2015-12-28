package robot

import (
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"net/http"
	"strconv"
	"time"
)

type WebServer struct {
	Dispatcher *mux.Router
	Server     *graceful.Server
	Urls       map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewWebServer(host string, port int) *WebServer {
	handler := mux.NewRouter()
	return &WebServer{
		Dispatcher: handler,
		Urls:       make(map[string]func(w http.ResponseWriter, r *http.Request)),
		Server: &graceful.Server{
			Server: &http.Server{Addr: host + ":" + strconv.Itoa(port), Handler: handler},
		},
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
	s.Server.ListenAndServe()
}

func (s *WebServer) Stop() {
	s.Server.Stop(1 * time.Second)
}
