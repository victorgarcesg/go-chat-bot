package auth

import (
	"encoding/json"
	"go-chat/persistence"
	"go-chat/settings"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func UserSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var user persistence.User
	json.NewDecoder(request.Body).Decode(&user)

	user.Password = GetHash([]byte(user.Password))

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

	jwtToken, err := GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
}

func GetHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte(settings.Cfg.Server.SecretKey))
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}

	return tokenString, nil
}
