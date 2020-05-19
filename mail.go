package main

import (
	"errors"
	"log"

	"gopkg.in/gomail.v2"
)

func MailSender(to string, token string, Name string, option string) error {

	if option == "signup" {
		m := gomail.NewMessage()
		m.SetHeader("From", "todoapi@gmail.com")
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Verification Code From Todo App")
		mailstring := "Hey \n" + Name + "\nWelocme To Todo App. Were Happy to see you " + "Here is Your Verification Code " + token
		m.SetBody("text/plain", mailstring)
		d := gomail.NewPlainDialer("localhost", 1025, "username", "password")
		if err := d.DialAndSend(m); err != nil {
			log.Println(err)
			return err
		}
		return nil

	} else if option == "reset" {
		m := gomail.NewMessage()
		m.SetHeader("From", "todoapi@gmail.com")
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Verification Code From Todo App")
		mailstring := "Hey \n" + Name + "\nHere, is your verification code " + token
		m.SetBody("text/plain", mailstring)
		d := gomail.NewPlainDialer("localhost", 1025, "username", "password")
		if err := d.DialAndSend(m); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("No Option Passed")

}
