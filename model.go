package main

import (
	"errors"
	"html"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name      string    `gorm:"size:255" json:"name"`
	Username  string    `gorm:"size:255; unique" json:"username"`
	Password  string    `gorm:"size:255" json:"password"`
	ID        uint32    `gorm:"primary_key;auto_increment:true" json:"id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
type Todo struct {
	Body      string `gorm:"size:255" json:"body"`
	ID        uint32 `gorm:"primary_key;auto_increment:true" json:"id"`
	UserID    uint32 `json:"userid"`
	Completed bool   `gorm:"default:false"  json:"completed"`
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPassword, password string) error {
	log.Println(hashedPassword)
	log.Println(password)
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func (u *User) HashBeforeSave() error {
	hashed, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}
func (u *User) Prepare() {
	u.ID = 0
	u.Name = html.EscapeString(u.Name)
	u.CreatedAt = time.Now()
}
func (u *User) SaveUser(db *gorm.DB) (*User, error) {

	var err error
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}
func (u *User) FindUserByID(db *gorm.DB, uid uint32) (*User, error) {
	var err error
	err = db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return u, err
}
func FindUserByname(db *gorm.DB, name string) (User, error) {
	var err error
	usr := User{}
	err = db.Debug().Model(User{}).Where("username = ?", name).Take(&usr).Error
	if err != nil {
		return User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return User{}, errors.New("User no found by this username")
	}
	return usr, nil
}

func (td *Todo) Prepare() {
	td.ID = 0
	td.Body = html.EscapeString(td.Body)
}
func (td *Todo) SaveTodo(db *gorm.DB) (*Todo, error) {
	var err error
	err = db.Debug().Create(&td).Error
	if err != nil {
		return &Todo{}, err
	}
	return td, nil
}
func (td *Todo) GetTodoByID(db *gorm.DB, id uint32) (*Todo, error) {
	var err error
	err = db.Debug().Model(&Todo{}).Where("id = ?", id).Take(&td).Error
	if err != nil {
		return &Todo{}, err
	}
	return td, nil
}
func (td *Todo) Getalltodo(db *gorm.DB, id uint32) (*[]Todo, error) {
	var err error
	todos := []Todo{}
	err = db.Debug().Model(&Todo{}).Where("user_id = ?", id).Find(&todos).Error
	if err != nil {
		return &[]Todo{}, err
	}
	return &todos, nil
}
func (td *Todo) DeleteTodo(db *gorm.DB, id uint32) (int64, error) {
	db = db.Debug().Model(&Todo{}).Where("id = ?", id).Take(&Todo{}).Delete(&Todo{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
