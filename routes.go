package main

func (s *Server) InitRoute() {
	s.Router.HandleFunc("/", JSONandCORS(s.Home)).Methods("GET")
	s.Router.HandleFunc("/signup", JSONandCORS(s.Signup)).Methods("POST")
	s.Router.HandleFunc("/signin", JSONandCORS(s.Signin)).Methods("POST")
	s.Router.HandleFunc("/addtodo", JSONandCORS(IsAuth(s.Addtodo))).Methods("POST")
	s.Router.HandleFunc("/dl/todo/{id}", JSONandCORS(IsAuth(s.Deletetodo))).Methods("DELETE")
}
