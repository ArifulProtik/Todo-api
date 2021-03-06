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
func Cors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
}

// Starting Middlewares

func JSONandCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}
func (s *Server) IsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := TokenValid(r)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Opps! Unauthorized",
			}, http.StatusUnauthorized)
			return
		}
		id, err := ExtractTokenID(r)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Opps! Unauthorized",
			}, http.StatusUnauthorized)
			return
		}
		getUser, err := FindUserByID(s.DB, id)
		if err != nil {
			log.Println(err)
			JSONWriter(w, data{
				"Error": "Internal Error",
			}, http.StatusUnauthorized)
			return
		}
		if getUser.IsActive != true {
			JSONWriter(w, data{"Error": "Email is Not Verified"}, http.StatusForbidden)
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

func (s *Server) Addtodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
func (s *Server) Gettodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id, err := ExtractTokenID(r)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	var td Todo
	todos, err := td.Getalltodo(s.DB, id)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	if len(*todos) == 0 {
		JSONWriter(w, data{
			"Error": "No Todos Yet!",
		}, http.StatusNotFound)
		return
	}
	JSONWriter(w, *todos, 200)
}

func (s *Server) Marktodo(w http.ResponseWriter, r *http.Request) {
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
	if todo.UserID != id {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	if !todo.Completed {
		todo, err = todo.UpdateCompleted(s.DB, true)
		if err != nil {
			JSONWriter(w, data{
				"Error": "Internal Error",
			}, http.StatusInternalServerError)
			return
		}
		JSONWriter(w, todo, 200)
		return
	}
	todo, err = todo.UpdateCompleted(s.DB, false)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, todo, 200)

}

func (s *Server) EditTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	requestTODO := Todo{}
	err = json.Unmarshal(body, &requestTODO)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found on Request",
		}, http.StatusUnprocessableEntity)
		return
	}
	if requestTODO.ID == 0 || requestTODO.Body == "" {
		JSONWriter(w, data{
			"Error": "Fields Can't be empty",
		}, http.StatusUnprocessableEntity)
		return
	}

	todo, err := requestTODO.GetTodoByID(s.DB, requestTODO.ID)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Todo Found",
		}, http.StatusNotFound)
		return
	}
	id, err := ExtractTokenID(r)
	if todo.UserID != id {
		JSONWriter(w, data{
			"Error": "Opps! Unauthorized",
		}, http.StatusUnauthorized)
		return
	}
	todo, err = todo.Update(s.DB, requestTODO.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, todo, 200)
}
