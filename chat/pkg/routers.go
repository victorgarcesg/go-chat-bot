package pkg

import (
	"encoding/json"
	"go-chat/auth"
	"net/http"

	"github.com/gorilla/mux"
)

func SetRouterHandlerFuncs(router *mux.Router, s *Server) {
	router.HandleFunc("/api/account/signup", auth.UserSignup).Methods("POST")
	router.HandleFunc("/api/account/login", auth.UserLogin).Methods("POST")
	router.HandleFunc("/api/rooms", func(rw http.ResponseWriter, r *http.Request) { getListRooms(s, rw, r) }).Methods("GET")
	router.HandleFunc("/home", serveHome)
	router.HandleFunc("/chat", chatRoom)
	router.HandleFunc("/signup", signupForm)
	router.HandleFunc("/", loginForm)
	router.HandleFunc("/ws", s.ServeWs)
}

var serveHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/home" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/home.html")
	})

var chatRoom = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/chat.html")
	})

var loginForm = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/login.html")
	})

var signupForm = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/signup" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/signup.html")
	})

func getListRooms(s *Server, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var hubsNames []string
	hubs := s.GetHubs()
	for name := range *hubs {
		hubsNames = append(hubsNames, name)
	}

	json, err := json.Marshal(hubsNames)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write(json)
}
