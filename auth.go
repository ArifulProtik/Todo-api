package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/dgrijalva/jwt-go"
)

const hello string = "key"

func CreateToken(user_id uint32) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = user_id
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(hello))

}

func TokenValid(r *http.Request) error {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(hello), nil
	})
	if err != nil {
		return err
	}
	_ = token
	return nil
}

func ExtractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractTokenID(r *http.Request) (uint32, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(hello), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(uid), nil
	}
	return 0, nil
}
func (s *Server) Signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found on Request",
		}, http.StatusUnprocessableEntity)
		return
	}
	if user.Email == "" || user.Password == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	getuser, err := FindUserByEmail(s.DB, user.Email)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No user found by Email " + user.Email,
		}, http.StatusNotFound)
		return
	}
	err = VerifyPassword(getuser.Password, user.Password)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Password incorrect",
		}, http.StatusNotFound)
		return
	}
	if getuser.IsActive != true {
		JSONWriter(w, data{
			"Error": "Email Is not Verified",
		}, http.StatusForbidden)
		return
	}
	token, err := CreateToken(getuser.ID)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{
		"token": token,
		"user": &User{
			Name:     getuser.Name,
			ID:       getuser.ID,
			Email:    getuser.Email,
			IsActive: getuser.IsActive,
			Avatar:   getuser.Avatar,
		},
	}, http.StatusAccepted)

}
func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found on Request",
		}, http.StatusUnprocessableEntity)
		return
	}
	user.Prepare()
	if user.Name == "" || user.Email == "" || user.Password == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	if err := checkmail.ValidateFormat(user.Email); err != nil {
		JSONWriter(w, data{
			"Error": "Email not valid",
		},
			http.StatusUnprocessableEntity)
		return
	}
	err = user.HashBeforeSave()
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	createduser, err := user.SaveUser(s.DB)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Username Already Exists",
		}, http.StatusConflict)
		return
	}
	otp, err := GenerateOTP(6)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	var Verify Verification
	Verify = Verification{
		UserID: createduser.ID,
		Token:  otp,
		Expiry: time.Now().Add(24 * time.Hour),
	}
	Verify, err = Verify.Save(s.DB)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error on Save OTP",
		},
			http.StatusInternalServerError)
		return
	}
	err = MailSender(createduser.Email, otp, createduser.Name, "signup")
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Accoun Created Successfully"}, 200)
}
