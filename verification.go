package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
)

type Verification struct {
	UserID uint32    `gorm:"not null" json:"userid"`
	Token  string    `gorm:"unique" json:"token"`
	Expiry time.Time `json:"expiry"`
}

const otpChars = "1234567890"

func GenerateOTP(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

//CRUD

func (v *Verification) Save(db *gorm.DB) (Verification, error) {
	var err error
	err = db.Debug().Create(&v).Error
	if err != nil {
		return Verification{}, err
	}
	return *v, nil
}
func GetByToken(db *gorm.DB, token string) (Verification, error) {
	var err error
	var v Verification
	err = db.Debug().Model(User{}).Where("token = ?", token).Take(&v).Error
	if err != nil {
		return Verification{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return Verification{}, errors.New("No Token Found")
	}
	return v, nil
}
func DeleteToken(db *gorm.DB, token string) (int64, error) {
	db = db.Debug().Model(&Verification{}).Where("token = ?", token).Take(&Verification{}).Delete(&Verification{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (s *Server) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	prep := strings.TrimSpace(param.Get("email"))
	if prep == "" {
		JSONWriter(w, data{
			"Error": "Email Missing",
		},
			http.StatusUnprocessableEntity)
		return
	}
	if err := checkmail.ValidateFormat(prep); err != nil {
		log.Println(prep)
		JSONWriter(w, data{
			"Error": "Email not valid",
		},
			http.StatusUnprocessableEntity)
		return
	}
	user, err := FindUserByEmail(s.DB, prep)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No EmailFound By " + prep,
		},
			http.StatusBadRequest)
		return
	}
	token, err := GenerateOTP(6)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error on generate OTP",
		},
			http.StatusInternalServerError)
		return
	}
	var Verify Verification
	Verify = Verification{
		UserID: user.ID,
		Token:  token,
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
	err = MailSender(user.Email, token, user.Name, "reset")
	if err != nil {
		// Deleting Token if Mailsending Encountered a Error
		effected, err := DeleteToken(s.DB, token)
		if err != nil {
			log.Println(err)
		}
		log.Println(effected)
		JSONWriter(w, data{
			"Error": "Internal Error On Mail Send",
		},
			http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{
		"Success": "A Verification Code sent to your Email",
	}, 200)

}

func (s *Server) ResetPassword(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	holder := Request{}
	err = json.Unmarshal(body, &holder)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found",
		}, http.StatusUnprocessableEntity)
		return
	}
	if holder.Token == "" || holder.Password == "" {
		JSONWriter(w, data{
			"Error": "Field cant be Empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	tokendata, err := GetByToken(s.DB, holder.Token)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Token is not valid",
		}, http.StatusBadRequest)
		return
	}
	password, err := Hash(holder.Password)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	hashed := string(password)
	err = ResetPass(s.DB, tokendata.UserID, hashed)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	rows, err := DeleteToken(s.DB, tokendata.Token)
	if err != nil {
		log.Println(err)
	}
	log.Println(rows)
	JSONWriter(w, data{"Success": "Password Has been updated"}, 200)
}
func (s *Server) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	prep := strings.TrimSpace(param.Get("token"))
	if prep == "" {
		JSONWriter(w, data{
			"Error": "token is missing",
		},
			http.StatusUnprocessableEntity)
		return
	}
	tokendata, err := GetByToken(s.DB, prep)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Token is not valid",
		}, http.StatusBadRequest)
		return
	}
	UpdateData := &User{
		IsActive: true,
	}
	err = UpdateUser(s.DB, *UpdateData, tokendata.UserID)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Email was verified Successfully"}, 200)
}

func (s *Server) ChangeMail(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Request Unprocessable",
		}, http.StatusUnprocessableEntity)
		return
	}
	holder := Request{}
	err = json.Unmarshal(body, &holder)
	if err != nil {
		JSONWriter(w, data{
			"Error": "No Json Found",
		}, http.StatusUnprocessableEntity)
		return
	}
	if holder.Email == "" || holder.Password == "" {
		JSONWriter(w, data{
			"Error": "Field Can't Be Empty",
		}, http.StatusUnprocessableEntity)
		return
	}
	if err := checkmail.ValidateFormat(holder.Email); err != nil {
		JSONWriter(w, data{
			"Error": "Email not valid",
		},
			http.StatusUnprocessableEntity)
		return
	}

	id, err := ExtractTokenID(r)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Unauthorized!!!",
		}, http.StatusUnauthorized)
		return
	}
	getUser, err := FindUserByID(s.DB, id)
	if err != nil {
		JSONWriter(w, data{"Error": "No User Found By This ID"}, http.StatusNotFound)
		return
	}
	err = VerifyPassword(getUser.Password, holder.Password)
	if err != nil {
		JSONWriter(w, data{"Error": "No User Found By This ID"}, http.StatusNotFound)
		return
	}
	if getUser.Email == holder.Email {
		JSONWriter(w, data{
			"Error": "Cant use Old Email",
		}, http.StatusInternalServerError)
		return
	}
	Updater := &User{
		Email:    holder.Email,
		IsActive: false,
	}
	err = UpdateMultiple(s.DB, Updater, id)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Email Already In Use",
		}, http.StatusInternalServerError)
		return
	}
	otp, err := GenerateOTP(6)
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	err = MailSender(holder.Email, otp, getUser.Name, "signup")
	if err != nil {
		JSONWriter(w, data{
			"Error": "Internal Error",
		}, http.StatusInternalServerError)
		return
	}
	JSONWriter(w, data{"Success": "Email Updated! Please Verify"}, 200)
}
