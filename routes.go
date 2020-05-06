package main

func (s *Server) InitRoute() {
	s.Router.HandleFunc("/", JSONandCORS(s.Home)).Methods("GET")
	s.Router.HandleFunc("/signup", JSONandCORS(s.Signup)).Methods("POST")
	s.Router.HandleFunc("/signin", JSONandCORS(s.Signin)).Methods("POST")
	s.Router.HandleFunc("/addtodo", JSONandCORS(IsAuth(s.Addtodo))).Methods("POST")
	s.Router.HandleFunc("/dl/todo/{id}", JSONandCORS(IsAuth(s.Deletetodo))).Methods("DELETE")
	s.Router.HandleFunc("/todos", JSONandCORS(IsAuth(s.Gettodo))).Methods("GET")

	// cors
	s.Router.HandleFunc("/", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/signup", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/signin", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/addtodo", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/dl/todo/{id}", Cors).Methods("OPTIONS")
	s.Router.HandleFunc("/todos", Cors).Methods("OPTIONS")
}
