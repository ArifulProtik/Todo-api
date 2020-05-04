package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type data map[string]interface{}

// Response
func JSONWriter(w http.ResponseWriter, data interface{}, statusCode int) {
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}

// Starting Middlewares

func JSONandCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}
func IsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := TokenValid(r)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Opps! Unauthorized",
			}, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// End Middleware

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	JSONWriter(w, data{
		"Messege": "Welcome To Todo",
	}, 200)
}

func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
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
	if user.Name == "" || user.Username == "" || user.Password == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
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
	} else {
		JSONWriter(w, data{
			"name":     createduser.Name,
			"username": createduser.Username,
		}, http.StatusCreated)
		return
	}

}
func (s *Server) Signin(w http.ResponseWriter, r *http.Request) {
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
	if user.Username == "" || user.Password == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	getuser, err := FindUserByname(s.DB, user.Username)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No user found by username " + user.Username,
		}, http.StatusNotFound)
		return
	}
	err = VerifyPassword(getuser.Password, user.Password)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Password incorrect",
		}, http.StatusNotFound)
		return
	} else {
		token, err := CreateToken(getuser.ID)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Internal Error",
			}, http.StatusInternalServerError)
		} else {
			JSONWriter(w, data{
				"name":     getuser.Name,
				"id":       getuser.ID,
				"username": getuser.Username,
				"token":    token,
			}, http.StatusAccepted)
		}

	}

}

func (s *Server) Addtodo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	todo := Todo{}
	err = json.Unmarshal(body, &todo)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Must Provide JSON",
		}, http.StatusUnprocessableEntity)
		return
	}
	if strings.TrimSpace(todo.Body) == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	id, err := ExtractTokenID(r)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	if id != todo.UserID {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	todo.Prepare()
	todos, err := todo.SaveTodo(s.DB)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, todos, http.StatusCreated)

}
func (s *Server) Deletetodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tid64, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		JSONWriter(w, data{
			"Error": err.Error(),
		}, http.StatusUnprocessableEntity)
		return
	}
	tid32 := uint32(tid64)

	id, err := ExtractTokenID(r)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	var td Todo
	todo, err := td.GetTodoByID(s.DB, tid32)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Todo Found",
		}, http.StatusNotFound)
		return
	}
	if todo == nil {
		JSONWriter(w, data{
			"Error": "No Todo Found By This ID",
		}, http.StatusNotFound)
		return
	}
	if id == todo.UserID {
		effect, err := todo.DeleteTodo(s.DB, tid32)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Internal Error",
			}, http.StatusInternalServerError)
			return
		}
		if effect == 0 {
			JSONWriter(w, data{
				"Error": "Something Went Wrong",
			}, http.StatusInternalServerError)
			return
		}
		JSONWriter(w, data{
			"Success": "Todo Deleted",
		}, http.StatusCreated)
		return
	}
	JSONWriter(w, data{
		"Error": "Opps! Unauthorized",
	}, http.StatusUnauthorized)

}
