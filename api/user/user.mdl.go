package user

import (
	"bytes"
	"fmt"
	"html/template"
	"mime/multipart"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type User struct {
	ID             uint   `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	FirstName      string `gorm:"column:firstname;not null" json:"firstname"`
	LastName       string `gorm:"column:lastname;not null" json:"lastname"`
	Email          string `gorm:"column:email;not null" json:"email"`
	VerifyCode     string `gorm:"column:verif_code;not null" json:"verif_code"`
	Attempts       uint   `gorm:"column:attempts"`
	IsVerified     bool   `gorm:"column:verif_status;default:false"`
	BirthDate      string `gorm:"column:birth_date;not null" json:"birth_date"`
	University     string `gorm:"column:university;not null" json:"university"`
	Phone          string `gorm:"column:phone;not null" json:"phone"`
	Password       string `gorm:"column:password;not null" json:"password"`
	Paiment_Status bool   `gorm:"column:paiment_status;not null" json:"paiment_status"`
	Paiment_Date   string `gorm:"column:paiment_date" json:"paiment_date"`
	LastLogin      string `gorm:"column:last_login" json:"last_login"`
	Role           string `gorm:"column:role;not null" json:"role"`
	SquadID        uint   `gorm:"column:squad_id" json:"squad_id"`
	CVURL          string `gorm:"column:cv_url" json:"cv_url"`

	gorm.Model
}

type LeaderLogIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LeaderLogedIn struct {
	Token string `gorm:"column:token" json:"token"`
}

// ForgotPasswordInput struct
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type ResetTokenUser struct {
	LeaderID   uint   `gorm:"column:user_id" json:"leader_id"`
	Email      string `gorm:"column:email;unique" json:"email"`
	ResetToken string `gorm:"column:token" json:"token"`
}

// ResetPasswordInput struct
type ResetPasswordInput struct {
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

type EmailData struct {
	URL       string
	FirstName string
	Subject   string
}
type File struct {
	File multipart.File `json:"file,omitempty" validate:"required"`
}

// hash password
func HashPassword(pass *string) {
	bytePass := []byte(*pass)
	hPass, _ := bcrypt.GenerateFromPassword(bytePass, bcrypt.DefaultCost)
	*pass = string(hPass)
}

// new user
func NewUser(db *gorm.DB, user User) (User, error) {
	return user, db.Create(&user).Error
}

// get all users
func GetAllUsers(db *gorm.DB) (users []User, err error) {
	return users, db.Find(&users).Error
}

// check user existence
func CheckUserExists(db *gorm.DB, id uint) bool {

	//init vars
	user := &User{}

	//check if user exist
	check := db.Where("id = ?", id).First(user)
	if check.Error != nil {
		return false
	}

	if check.RowsAffected == 0 {
		return false
	} else {
		return true
	}
}

// Check user in Squad
// Check user existance in another squad
func CheckUserInSquad(db *gorm.DB, email string) bool {

	//init vars
	user := &User{}
	//check if user exist
	check := db.Where("email = ?", email).First(user)
	if check.Error != nil {
		return false
	}

	if check.RowsAffected == 0 {
		return false
	} else {
		return true
	}

}

// update user
func UpdateUser(db *gorm.DB, user User) error {
	return db.Where("id=?", user.ID).Updates(&user).Error
}

// Delete user
func DeleteUser(db *gorm.DB, user_id uint) error {
	return db.Where("id = ?", user_id).Delete(&User{}).Error
}

// get user by email
func GetUserByEmail(db *gorm.DB, email string) (user User, err error) {
	return user, db.First(&user, "email=?", email).Error
}

// Get user by id
func GetUserByID(db *gorm.DB, user_id uint) (user User, err error) {
	return user, db.Where("id = ?", user_id).First(&user).Error
}

// Get user by Role
func GetUsersByRole(db *gorm.DB, role_name string) (users []User, err error) {
	return users, db.Where("role = ?", role_name).Find(&users).Error
}

// Get members by squadID
func GetMembersBySquadID(db *gorm.DB, squad_id uint) (users []User, err error) {
	return users, db.Where("squad_id = ?", squad_id).Find(&users).Error
}

// compare two passwords
func ComparePasswords(dbpass, pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(dbpass), []byte(pass)) == nil
}

// Verify email
func VerifyEmail(db *gorm.DB, code string) (user User, err error) {
	return user, db.Where("verif_code = ?", code).First(&user).Error
}

// Send Email
func SendGomail(subject, email, templatePath string, user User) {

	// Get the HTML template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		fmt.Println("Failed to parse HTML template:", err)
		return
	}

	var body bytes.Buffer
	if err := t.Execute(&body, struct{ FirstName, LastName, VerifyCode string }{FirstName: user.FirstName, LastName: user.LastName, VerifyCode: user.VerifyCode}); err != nil {
		fmt.Println("Failed to execute template:", err)
		return
	}

	// Send With Gomail
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_SENDER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(os.Getenv("EMAIL_SMTP_SERVER"), 587, os.Getenv("EMAIL_SENDER"), "lbyruvvlgonrndxv")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Failed to send email:", err)
		return
	}

}

func SendForgetGomail(data *EmailData, email, templatePath string, user User) {

	// Get the HTML template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		fmt.Println("Failed to parse HTML template:", err)
		return
	}

	var body bytes.Buffer
	if err := t.Execute(&body, struct{ FirstName, LastName, URL string }{FirstName: user.FirstName, LastName: user.LastName, URL: data.URL}); err != nil {
		fmt.Println("Failed to execute template:", err)
		return
	}

	// Send With Gomail
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_SENDER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", body.String())

	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(os.Getenv("EMAIL_SMTP_SERVER"), 587, os.Getenv("EMAIL_SENDER"), "lbyruvvlgonrndxv")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Failed to send email:", err)
		return
	}

}

// Send Validation Email
func SendValidationGomail(subject, email, templatePath string, user User) {

	// Get the HTML template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		fmt.Println("Failed to parse HTML template:", err)
		return
	}

	var body bytes.Buffer
	if err := t.Execute(&body, struct{ FirstName, LastName string }{FirstName: user.FirstName, LastName: user.LastName}); err != nil {
		fmt.Println("Failed to execute template:", err)
		return
	}

	// Send With Gomail
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_SENDER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(os.Getenv("EMAIL_SMTP_SERVER"), 587, os.Getenv("EMAIL_SENDER"), "lbyruvvlgonrndxv")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Failed to send email:", err)
		return
	}

}
