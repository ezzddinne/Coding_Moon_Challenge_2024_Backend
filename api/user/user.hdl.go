package user

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/ezzddinne/CMC_BACKEND/middleware_reset"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Database struct {
	DB       *gorm.DB
	Enforcer *casbin.Enforcer
}

func (db Database) NewLeader(ctx *gin.Context) {

	// init vars
	var leader User
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&leader); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check values validity
	if empty_reg.MatchString(leader.FirstName) || empty_reg.MatchString(leader.LastName) || empty_reg.MatchString(leader.Email) || empty_reg.MatchString(leader.University) || empty_reg.MatchString(leader.Phone) || empty_reg.MatchString(leader.BirthDate) || empty_reg.MatchString(leader.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	//hash password
	HashPassword(&leader.Password)

	//Create verif code
	token := uuid.New().String()

	//init new leader
	new_leader := User{
		FirstName:      leader.FirstName,
		LastName:       leader.LastName,
		Email:          leader.Email,
		VerifyCode:     token,
		IsVerified:     false,
		University:     leader.University,
		Phone:          leader.Phone,
		BirthDate:      leader.BirthDate,
		Password:       leader.Password,
		Paiment_Status: false,
		Paiment_Date:   "0",
		Role:           "leader",
	}

	//Create leader
	new_leader_created, err := NewUser(db.DB, new_leader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Send email with code
	subject := "Coding Moon Community Want To Say Hi !"

	// Send Email
	SendGomail(subject, new_leader_created.Email, "api/user/Registration.html", new_leader_created)

	//leader created successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Leader created successfully"})

}

// Verify user
func (db Database) handleEmailVerification(ctx *gin.Context) {

	// init vars
	var verificationCode struct {
		VerifyCode string `json:"verif_code"`
	}
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&verificationCode); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check values validity
	if empty_reg.MatchString(verificationCode.VerifyCode) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	// get the email from the path
	email := ctx.Param("email")

	dbUser, err := GetUserByEmail(db.DB, email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Attempt to verify the email with the provided verification code
	_, err = VerifyEmail(db.DB, verificationCode.VerifyCode)
	if err != nil {
		// Increment the failed verification attempts
		dbUser.Attempts++

		// Update user in the database
		if err := UpdateUser(db.DB, dbUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to update user"})
			return
		}

		// Check if the user has exceeded the maximum allowed attempts
		if dbUser.Attempts >= 5 {
			// Delete the user from the database due to exceeding attempts
			if err := DeleteUser(db.DB, dbUser.ID); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user"})
				return
			}
			// Respond with failure message and delete user
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Exceeded maximum verification attempts"})
			return
		}

		// Respond with failure message and remaining attempts
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid verification code"})
		return
	}

	// Reset failed attempts since verification was successful
	dbUser.Attempts = 0

	// Update user's verification status to true
	dbUser.IsVerified = true

	// Update user in the database
	if err := UpdateUser(db.DB, dbUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to update user"})
		return
	}

	// Respond with success message
	ctx.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})

}

func (db Database) SignInLeader(ctx *gin.Context) {

	//init vars
	var leader_login LeaderLogIn
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&leader_login); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	// check field validity
	if empty_reg.MatchString(leader_login.Email) || empty_reg.MatchString(leader_login.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	//check if email exists ==> user
	dbLeader, err := GetUserByEmail(db.DB, leader_login.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No Such User Found"})
		return
	}

	// Verify if the user is verified
	if dbLeader.IsVerified {

		// update last login
		dbLeader.LastLogin = time.Now().Format("2006-01-02 15:04:05")

		// update user
		if err := UpdateUser(db.DB, dbLeader); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//compare password
		if isTrue := ComparePasswords(dbLeader.Password, leader_login.Password); isTrue {

			//generate token
			token := middleware.GenerateToken(dbLeader.ID, dbLeader.SquadID, dbLeader.Role)
			ctx.JSON(http.StatusOK, LeaderLogedIn{Token: token})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "password not matched"})
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "This user is not verified"})
	}

}

// Get all users
func (db Database) GetAllUsers(ctx *gin.Context) {

	users, err := GetAllUsers(db.DB)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//return users
	ctx.JSON(http.StatusOK, users)

}

// get user by id
func (db Database) GetUserByID(ctx *gin.Context) {

	//get the user id
	user_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//get user by id
	user, err := GetUserByID(db.DB, uint(user_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//response
	ctx.JSON(http.StatusOK, user)
}

// get user by id to merge squad
func (db Database) GetUserByIDFront(ctx *gin.Context) {

	//get the user id
	session := middleware.ExtractTokenValues(ctx)

	//get user by id
	user, err := GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//response
	ctx.JSON(http.StatusOK, user)
}

// get User by email
func (db Database) GetUserByEmail(ctx *gin.Context) {

	// get the email from the path
	email := ctx.Param("email")

	// get yser by email
	user, err := GetUserByEmail(db.DB, email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	// return the data
	ctx.JSON(http.StatusOK, user)
}

// get users by role
func (db Database) GetUsersByRole(ctx *gin.Context) {

	// get value from the path
	role := ctx.Param("role")

	// get users by role
	users, err := GetUsersByRole(db.DB, role)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	// return users
	ctx.JSON(http.StatusOK, users)
}

func (db Database) DeleteUser(ctx *gin.Context) {

	// get id from path
	user_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//delete the user
	//Check
	if err = DeleteUser(db.DB, uint(user_id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Deleted successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// change paiment status
func (db Database) ChangePaimentStatus(ctx *gin.Context) {

	// get id value from path
	user_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	paiment_status := User{
		ID:             uint(user_id),
		Paiment_Status: true,
		Paiment_Date:   time.Now().Format("2006-01-02 15:04:05"),
	}

	//update paiment
	//check update
	if err := UpdateUser(db.DB, paiment_status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//updated successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Paiment Status changed successfully"})
}

// get users by squad ID
func (db Database) GetUsersBySquadID(ctx *gin.Context) {

	// get the value from the path
	squad_id, err := strconv.Atoi(ctx.Param("squad_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// get the users by squad ID
	users, err := GetMembersBySquadID(db.DB, uint(squad_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// return members
	ctx.JSON(http.StatusOK, users)
}

func (db Database) ForgetPassword(ctx *gin.Context) {

	//init vars
	var userrsp ForgotPasswordInput
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&userrsp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check values validity
	if empty_reg.MatchString(userrsp.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "field invalid"})
		return
	}

	//verify that the user exist by email
	dbLeader, err := GetUserByEmail(db.DB, userrsp.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "User with this email dosen't exist"})
		return
	}

	//generate token
	token := middleware_reset.GenerateResetToken(dbLeader.ID, dbLeader.SquadID, dbLeader.Role)

	ctx.JSON(http.StatusOK, ResetTokenUser{LeaderID: dbLeader.ID, Email: dbLeader.Email, ResetToken: token})

	ctx.JSON(http.StatusOK, gin.H{"message": "Request sended successfully"})

	// Send Emails
	emailData := EmailData{
		URL:     "/resetpassword/" + token,
		Subject: "Reset Your Password",
	}

	SendForgetGomail(&emailData, dbLeader.Email, "api/user/Reset_password.html", dbLeader)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Check your mail please"})
}

func (db Database) ResetPassword(ctx *gin.Context) {

	//init vars
	var reset ResetPasswordInput
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&reset); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user := middleware_reset.ExtractResetTokenValues(ctx)

	// check if the time expired
	if time.Now().After(user.ExpiresAt) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Time has expired"})
		return
	}

	// check values validity
	if empty_reg.MatchString(reset.Password) || empty_reg.MatchString(reset.PasswordConfirm) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "field invalid"})
		return
	}

	//compare the password and the confirmation
	if reset.Password != reset.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "The password dosen't match"})
		return
	}

	HashPassword(&reset.Password)

	// get user by id
	dbLeader, err := GetUserByID(db.DB, user.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	dbLeader.Password = reset.Password

	// update user
	if err := UpdateUser(db.DB, dbLeader); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})

}

func (db Database) LeaderFileUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		//upload
		formFile, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				MediaDto{
					StatusCode: http.StatusInternalServerError,
					Message:    "error",
					Data:       map[string]interface{}{"data": "Select a file to upload"},
				})
			return
		}

		// get id from path
		user_email := c.Param("email")

		// get user by ID
		dbLeader, err := GetUserByEmail(db.DB, user_email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		uploadUrl, err := NewMediaUpload().FileUpload(File{File: formFile})
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				MediaDto{
					StatusCode: http.StatusInternalServerError,
					Message:    "error",
					Data:       map[string]interface{}{"data": "Error uploading file"},
				})
			return
		}

		c.JSON(
			http.StatusOK,
			MediaDto{
				StatusCode: http.StatusOK,
				Message:    "success",
				Data:       map[string]interface{}{"data": uploadUrl + " : " + dbLeader.Email},
			})

		//init vars
		dbLeader.CVURL = uploadUrl

		// update squad
		if err := UpdateUser(db.DB, dbLeader); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Leader cv added successfully"})
	}

}
