package main

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/twinj/uuid"
)

func GenUID() string {
	u := uuid.NewV4()
	return u.String()
}
func (s *Server) UploadProfile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	id, err := ExtractTokenID(r)
	if err != nil {
		JSONWriter(w, data{"Error": "UnAuthorized"}, http.StatusUnauthorized)
		return
	}
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	size := handler.Size
	if size > 1000000 {
		JSONWriter(w, data{"Error": "More then 1 MB size of Images not allowed"}, http.StatusUnprocessableEntity)
		return
	}
	if !strings.Contains(handler.Header.Get("Content-Type"), "image") {
		JSONWriter(w, data{"Error": "Please Upload a Image"}, http.StatusUnprocessableEntity)
		return

	}
	uid := GenUID()
	filename := uid + ".jpg"
	f, err := os.OpenFile("./propic/"+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		JSONWriter(w, data{"Error": "Internal Error"}, http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	url := "http://localhost:8081/images/" + filename
	UpdateData := &User{
		Avatar: url,
	}
	err = UpdateUser(s.DB, *UpdateData, id)
	if err != nil {
		JSONWriter(w, data{"Error": "Internal Error"}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Uploaded!!!"}, 200)
}
