package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// Wraps server muxer, dynamic map of handlers, and listen port.
type WebServer struct {
	Dispatcher *mux.Router
	Urls       map[string]func(w http.ResponseWriter, r *http.Request)
	Port       string
}

func (s *WebServer) Start() {
	http.ListenAndServe(":"+s.Port, s.Dispatcher)
}

// Initialize Dispatcher's routes.
func (s *WebServer) InitDispatch() {
	d := s.Dispatcher

	// Add handler to server's map.
	d.HandleFunc("/register/{name}", func(w http.ResponseWriter, r *http.Request) {
		//somewhere somehow you create the handler to be used; i'll just make an echohandler
		vars := mux.Vars(r)
		name := vars["name"]

		s.AddFunction(w, r, name)
	}).Methods("GET")

	d.HandleFunc("/destroy/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		s.Destroy(name)
	}).Methods("GET")

	d.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		//Lookup handler in map and call it, proxying this writer and request
		vars := mux.Vars(r)
		name := vars["name"]

		s.ProxyCall(w, r, name)
	}).Methods("GET")
}

func (s *WebServer) Destroy(fName string) {
	s.Urls[fName] = nil //remove handler
}

func (s *WebServer) ProxyCall(w http.ResponseWriter, r *http.Request, fName string) {
	if s.Urls[fName] != nil {
		s.Urls[fName](w, r) //proxy the call
	}
}

func (s *WebServer) AddFunction(w http.ResponseWriter, r *http.Request, fName string) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from" + fName))
	}

	s.Urls[fName] = f // Add the handler to our map
}

func NewWebServer(port int) *WebServer {
	return &WebServer{
		Port:       strconv.Itoa(port),
		Dispatcher: mux.NewRouter(),
		Urls:       make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
}

// // May the signal never stop.
// func main() {
// 	server := NewWebServer(3000)
// 	server.InitDispatch()
// 	server.Start() //Launch server; blocks goroutine.
// }
