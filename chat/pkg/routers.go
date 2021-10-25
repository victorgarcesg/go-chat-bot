package pkg

import (
	"encoding/json"
	"go-chat/auth"
	"go-chat/persistence"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func SetRouterHandlerFuncs(router *mux.Router, s *Server) {
	router.HandleFunc("/api/account/signup", UserSignup).Methods("POST")
	router.HandleFunc("/api/account/login", UserLogin).Methods("POST")
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
	for name, _ := range *hubs {
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

func UserSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user persistence.User
	json.NewDecoder(request.Body).Decode(&user)

	user.Password = auth.GetHash([]byte(user.Password))

	result := persistence.DB.Create(&user)

	if result.Error != nil {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte("user not created"))
		return
	}

	json.NewEncoder(response).Encode(result)
}

func UserLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user persistence.User
	var dbUser persistence.User
	json.NewDecoder(request.Body).Decode(&user)
	persistence.DB.Where(&persistence.User{Username: user.Username}).First(&dbUser)

	if (persistence.User{}) == dbUser {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte("user not found"))
		return
	}

	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}

	jwtToken, err := auth.GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
}
