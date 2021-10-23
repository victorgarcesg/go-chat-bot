package persistence

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = []byte("veryC0mpl3j0")

func GetHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}

	return tokenString, nil
}

func userSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user User
	json.NewDecoder(request.Body).Decode(&user)
	user.Password = GetHash([]byte(user.Password))
	result := DB.Create(&user)
	if result.Error != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte("user not created"))
		return
	}

	json.NewEncoder(response).Encode(result)
}

func userLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&user)
	DB.Where(&User{Username: user.Username}).First(&dbUser)

	if (User{}) == dbUser {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte("user not found"))
		return
	}
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}

	jwtToken, err := GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
}
