package squad

import (
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/api/user"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Database struct {
	DB       *gorm.DB
	Enforcer *casbin.Enforcer
}

func (db Database) CreateSquad(ctx *gin.Context) {

	//init vars
	var squad Squad
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	//check json validity
	if err := ctx.ShouldBindJSON(&squad); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check fields
	if empty_reg.MatchString(squad.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	// get values from middleware
	session := middleware.ExtractTokenValues(ctx)

	dbLeader, err := user.GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check user exist
	// if exist can't create another Squad
	if CheckUserCreateSquad(db.DB, dbLeader.ID) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "User already create a squad"})
	} else {

		//init vars
		var intIDs []int32
		intIDs = append(intIDs, int32(dbLeader.ID))

		//init new squad
		new_squad := Squad{
			Name:         squad.Name,
			CreatedBy:    dbLeader.ID,
			SquadMembers: intIDs,
		}

		//create new squad
		new_squad_created, err := NewSquad(db.DB, new_squad)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//squad created succsessfully
		ctx.JSON(http.StatusOK, gin.H{"message": "Squad created successfully"})

		// update user squad id
		dbLeader.SquadID = new_squad_created.ID

		// update user
		if err := user.UpdateUser(db.DB, dbLeader); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		subject := "Payment Process"

		// Send Email
		user.SendValidationGomail(subject, dbLeader.Email, "api/user/Validation.html", dbLeader)

		ctx.JSON(http.StatusOK, gin.H{"message": "Mail sended successfully"})

	}

}

// Get all squads
func (db Database) GetAllSquads(ctx *gin.Context) {

	squads, err := GetAllSquads(db.DB)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//return squads
	ctx.JSON(http.StatusOK, squads)

}

func (db Database) GetSquadByID(ctx *gin.Context) {

	//get the squad id
	squad_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//get squad by id
	squad, err := GetSquadByID(db.DB, uint(squad_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//response
	ctx.JSON(http.StatusOK, squad)
}

func (db Database) GetSquadByIDFront(ctx *gin.Context) {

	// get session value
	session := middleware.ExtractTokenValues(ctx)

	//get leader by id
	dbLeader, err := user.GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//get squad by id
	squad, err := GetSquadByID(db.DB, dbLeader.SquadID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//response
	ctx.JSON(http.StatusOK, squad)
}

func (db Database) GetSquadByEmail(ctx *gin.Context) {

	//get the squad email
	squad_email := ctx.Param("email")

	//get squad by email
	squad, err := GetSquadByEmail(db.DB, squad_email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//response
	ctx.JSON(http.StatusOK, squad)
}

func (db Database) DeleteSquad(ctx *gin.Context) {

	// get the session values
	session := middleware.ExtractTokenValues(ctx)

	//get user
	dbLeader, err := user.GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	dbSquad, err := GetSquadByID(db.DB, dbLeader.SquadID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// delete squad
	if err := DeleteSquad(db.DB, dbLeader.SquadID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Squad Deleted Successfully"})

	// Delete all users
	for _, id := range dbSquad.SquadMembers {

		if id == int32(dbLeader.ID) {
			continue
		} else {
			// delete user
			if err := user.DeleteUser(db.DB, uint(id)); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
				return
			}
		}

	}
}

func (db Database) ImageUpload() gin.HandlerFunc {
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

		//get values
		session := middleware.ExtractTokenValues(c)

		// get user by ID
		dbUser, err := user.GetUserByID(db.DB, session.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// get Squad by ID
		dbSquad, err := GetSquadByID(db.DB, dbUser.SquadID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		uploadUrl, err := NewMediaUpload().ImageUpload(File{File: formFile})
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
				Data:       map[string]interface{}{"data": uploadUrl + " : " + dbSquad.Name},
			})

		dbSquad.LogoURL = uploadUrl

		// update squad
		if err := UpdateSquad(db.DB, dbSquad); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Squad updated successfully"})

	}

}

func (db Database) FileUpload() gin.HandlerFunc {
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

		//get values
		session := middleware.ExtractTokenValues(c)

		// get user by ID
		dbUser, err := user.GetUserByID(db.DB, session.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// get Squad by ID
		dbSquad, err := GetSquadByID(db.DB, dbUser.SquadID)
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
				Data:       map[string]interface{}{"data": uploadUrl + " : " + dbSquad.Name},
			})

		//init vars
		var stringCVs []string
		stringCVs = append(stringCVs, uploadUrl)
		dbSquad.CvURLS = append(dbSquad.CvURLS, stringCVs...)

		// update squad
		if err := UpdateSquad(db.DB, dbSquad); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Squad updated successfully"})
	}

}

func (db Database) AddMember(ctx *gin.Context) {

	// init vars
	var vuser user.User
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&vuser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check values validity
	if empty_reg.MatchString(vuser.FirstName) || empty_reg.MatchString(vuser.LastName) || empty_reg.MatchString(vuser.Email) || empty_reg.MatchString(vuser.University) || empty_reg.MatchString(vuser.Phone) || empty_reg.MatchString(vuser.BirthDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	if user.CheckUserInSquad(db.DB, vuser.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "User exist in other squad"})
		return
	} else {

		//get values from session
		session := middleware.ExtractTokenValues(ctx)

		// get leader by id
		dbLeader, err := user.GetUserByID(db.DB, session.UserID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//init new member
		new_member := user.User{
			FirstName:      vuser.FirstName,
			LastName:       vuser.LastName,
			Email:          vuser.Email,
			University:     vuser.University,
			Phone:          vuser.Phone,
			BirthDate:      vuser.BirthDate,
			Paiment_Status: false,
			Role:           "member",
			SquadID:        dbLeader.SquadID,
		}

		//add a member to squad
		new_member_created, err := user.NewUser(db.DB, new_member)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//user added successfully
		ctx.JSON(http.StatusOK, gin.H{"message": "member added successfully"})

		//get squad by id
		dbSquad, err := GetSquadByID(db.DB, dbLeader.SquadID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//init vars
		var intIDs []int32
		intIDs = append(intIDs, int32(new_member_created.ID))
		dbSquad.SquadMembers = append(dbSquad.SquadMembers, intIDs...)

		// update squad
		if err := UpdateSquad(db.DB, dbSquad); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		//member added to slice successfully
		ctx.JSON(http.StatusOK, gin.H{"message": "member added to slice successfully"})

		subject := "Payment Process"

		// Send Email
		user.SendValidationGomail(subject, new_member.Email, "api/user/Validation.html", new_member_created)

		ctx.JSON(http.StatusOK, gin.H{"message": "Mail sended successfully"})
	}

}

// Update Squad Name
func (db Database) UpdateName(ctx *gin.Context) {

	// init vars
	var usquad Squad
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// unmarshal sent json
	if err := ctx.ShouldBindJSON(&usquad); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check fields
	if empty_reg.MatchString(usquad.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	// get session value
	session := middleware.ExtractTokenValues(ctx)

	//get leader by id
	dbLeader, err := user.GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//get squad by id
	dbSquad, err := GetSquadByID(db.DB, dbLeader.SquadID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	dbSquad.Name = usquad.Name

	// update squad
	if err := UpdateSquad(db.DB, dbSquad); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Squad Name changed successfully"})

}

// get squad logo
func (db Database) GetSquadLogo(ctx *gin.Context) {

	// get values from session
	session := middleware.ExtractTokenValues(ctx)

	// get user by ID
	dbUser, err := user.GetUserByID(db.DB, session.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// get Squad by ID
	dbSquad, err := GetSquadByID(db.DB, dbUser.SquadID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": dbSquad.LogoURL})
}
