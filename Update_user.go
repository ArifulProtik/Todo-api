package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func (s *Server) UpdateName(w http.ResponseWriter, r *http.Request) {
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
	if user.Name == "" {
		JSONWriter(w, data{
			"Error": "Name Field Is Empty",
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
	// Addiong A Layer So that no can Pass All Field through Updater interface
	RequestData := &User{
		Name: user.Name,
	}
	err = UpdateUser(s.DB, *RequestData, id)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Name Updated Successfully"}, 200)
}
func (s *Server) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	type RequestData struct {
		OldPass string `json:"oldpass"`
		NewPass string `json:"newpass"`
	}
	req := RequestData{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found on Request",
		}, http.StatusUnprocessableEntity)
		return
	}
	if req.NewPass == "" || req.OldPass == "" {
		JSONWriter(w, data{
			"Error": "Name Field Is Empty",
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
	getUser, err := FindUserByID(s.DB, id)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	err = VerifyPassword(getUser.Password, req.OldPass)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Old Password Didn't match",
		}, http.StatusUnauthorized)
		return
	}
	SendData := &User{
		Password: req.NewPass,
	}
	err = UpdateUser(s.DB, *SendData, id)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Password Updated!"}, 200)
}
