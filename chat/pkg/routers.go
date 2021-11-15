package pkg

import (
	"encoding/json"
	"fmt"
	"go-chat/auth"
	"go-chat/settings"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

func SetRouterHandlerFuncs(router *mux.Router, s *Server) {
	router.HandleFunc("/", loginForm)
	router.HandleFunc("/ws", s.ServeWs)
	router.HandleFunc("/home", serveHome)
	router.HandleFunc("/chat", chatRoom)
	router.HandleFunc("/signup", signupForm)

	api := router.PathPrefix("/api").Subrouter()

	account := api.PathPrefix("/account").Subrouter()
	account.HandleFunc("/signup", auth.UserSignup).Methods("POST")
	account.HandleFunc("/login", auth.UserLogin).Methods("POST")

	rooms := api.PathPrefix("/rooms").Subrouter()
	rooms.HandleFunc("", func(rw http.ResponseWriter, r *http.Request) { getListRooms(s, rw, r) }).Methods("GET")
	rooms.HandleFunc("/{room}/messages", getRoomMessages).Methods("GET")
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
	if !isAuthorized(request, response) {
		return
	}

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

func isAuthorized(request *http.Request, response http.ResponseWriter) bool {
	if request.Header["Authorization"] != nil {
		token, err := jwt.Parse(request.Header["Authorization"][0], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				msg := "Error in JWT token validation"
				return nil, fmt.Errorf("%s", msg)
			}
			return []byte(settings.Cfg.Server.SecretKey), nil
		})

		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			response.Write([]byte(err.Error()))
			return false
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return true
		} else {
			response.WriteHeader(http.StatusUnauthorized)
			response.Write([]byte("Not Authorized"))
			return false
		}
	} else {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte("Not Authorized"))
		return false
	}
}

func getRoomMessages(response http.ResponseWriter, request *http.Request) {
	if !isAuthorized(request, response) {
		return
	}

	response.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(request)
	param := vars["room"]
	if param == "" {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"Missing proper query string parameter"}`))
		return
	}

	json, err := json.Marshal(RoomsMessages["#"+param])
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write(json)
}
