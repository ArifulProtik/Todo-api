package main

import (
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"html"
	"log"
	"time"
)

type User struct {
	Name      string    `gorm:"size:255" json:"name"`
	Email     string    `gorm:"size:255; unique" json:"email"`
	Password  string    `gorm:"size:255" json:"password"`
	ID        uint32    `gorm:"primary_key;auto_increment:true" json:"id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	IsActive  bool      `gorm:"default:false" json:"isactive"`
	Avatar    string    `gorm:"default:'https://www.pinclipart.com/picdir/middle/287-2871700_avatar-placeholder-clipart.png'" json:"avater"`
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
func FindUserByID(db *gorm.DB, uid uint32) (User, error) {
	var err error
	var u User
	err = db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return User{}, errors.New("User Not Found")
	}
	return u, err
}
func FindUserByEmail(db *gorm.DB, email string) (User, error) {
	var err error
	usr := User{}
	err = db.Debug().Model(User{}).Where("email = ?", email).Take(&usr).Error
	if err != nil {
		return User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return User{}, errors.New("User no found by this username")
	}
	return usr, nil
}
func ResetPass(db *gorm.DB, id uint32, password string) error {
	usr := User{}
	err := db.Debug().Model(User{}).Where("id = ?", id).Update("password", password).Take(&usr).Error
	if err != nil {
		return err
	}
	return nil
}
func UpdateUser(db *gorm.DB, user User, id uint32) error {
	usr := User{}
	log.Println()
	err := db.Debug().Model(User{}).Where("id = ?", id).Updates(user).Take(&usr).Error
	if err != nil {
		return err
	}
	return nil
}
func UpdateMultiple(db *gorm.DB, user *User, id uint32) error {
	usr := User{}

	err := db.Debug().Model(User{}).Where("id = ?", id).Updates(map[string]interface{}{"email": user.Email, "is_active": false}).Take(&usr).Error
	if err != nil {
		return err
	}
	return nil
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
func (td *Todo) UpdateCompleted(db *gorm.DB, value bool) (*Todo, error) {
	db = db.Debug().Model(&td).UpdateColumn("completed", value)
	if db.Error != nil {
		return &Todo{}, db.Error
	}
	err := db.Debug().Model(&Todo{}).Where("id = ?", td.ID).Take(&td).Error
	if err != nil {
		return &Todo{}, err
	}
	return td, nil
}
func (td *Todo) Update(db *gorm.DB, body string) (*Todo, error) {
	var err error
	err = db.Debug().Model(&Todo{}).Where("id = ?", td.ID).Updates(Todo{Body: body}).Error
	if db.Error != nil {
		return &Todo{}, db.Error
	}
	err = db.Debug().Model(&Todo{}).Where("id = ?", td.ID).Take(&td).Error
	if err != nil {
		return &Todo{}, err
	}
	return td, nil
}
