package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

func (s *Server) Initialize() {
	var err error
	connection := fmt.Sprintf("host=127.0.0.1 port=5432 user=protik dbname=todo password=112233 sslmode=disable")
	s.DB, err = gorm.Open("postgres", connection)
	if err != nil {
		fmt.Println("DB not connected")
		log.Fatal(err)
		os.Exit(1)
	}
	s.DB.Debug().AutoMigrate(
		User{},
		Todo{},
		Verification{},
	)
	s.Router = mux.NewRouter()
	s.InitRoute()
}
func main() {
	server := Server{}
	server.Initialize()
	fmt.Println("Server Starting on Port 8081")
	log.Fatal(http.ListenAndServe(":8081", server.Router))

}
