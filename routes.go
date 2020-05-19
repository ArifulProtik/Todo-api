package main

import "net/http"

func (s *Server) InitRoute() {
	s.Router.HandleFunc("/", JSONandCORS(s.Home)).Methods("GET")
	s.Router.HandleFunc("/signup", JSONandCORS(s.Signup)).Methods("POST")
	s.Router.HandleFunc("/signin", JSONandCORS(s.Signin)).Methods("POST")
	s.Router.HandleFunc("/addtodo", JSONandCORS(s.IsAuth(s.Addtodo))).Methods("POST")
	s.Router.HandleFunc("/updatename", JSONandCORS(s.IsAuth(s.UpdateName))).Methods("POST")
	s.Router.HandleFunc("/updatepass", JSONandCORS(s.IsAuth(s.UpdatePassword))).Methods("POST")
	s.Router.HandleFunc("/updatemail", JSONandCORS(s.IsAuth(s.ChangeMail))).Methods("POST")
	s.Router.HandleFunc("/updateprofilepic", JSONandCORS(s.IsAuth(s.UploadProfile))).Methods("POST")

	s.Router.HandleFunc("/dl/todo/{id}", JSONandCORS(s.IsAuth(s.Deletetodo))).Methods("DELETE")
	s.Router.HandleFunc("/todos", JSONandCORS(s.IsAuth(s.Gettodo))).Methods("GET")
	s.Router.HandleFunc("/check/todo/{id}", JSONandCORS(s.IsAuth(s.Marktodo))).Methods("PUT")
	s.Router.HandleFunc("/update/todo/{id}", JSONandCORS(s.IsAuth(s.EditTodo))).Methods("PUT")
	s.Router.HandleFunc("/forgetpassword", JSONandCORS(s.ForgetPassword)).Queries("email", "{email}").Methods("GET")
	s.Router.HandleFunc("/validate", JSONandCORS(s.ValidateEmail)).Queries("token", "{token}").Methods("GET")
	s.Router.HandleFunc("/resetpassword", JSONandCORS(s.ResetPassword)).Methods("POST")
	s.Router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("propic"))))

	// cors
	s.Router.HandleFunc("/", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/signup", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/signin", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/addtodo", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/dl/todo/{id}", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/todos", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/check/todo/{id}", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/update/todo/{id}", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/forgetpassword", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/resetpassword", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/updatename", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/updatepass", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/updatemail", Cors).Methods("OPTIONS")

}
