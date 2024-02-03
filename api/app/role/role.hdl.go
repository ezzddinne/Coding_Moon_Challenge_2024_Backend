package role

import (
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Database struct {
	DB       *gorm.DB
	Enforcer *casbin.Enforcer
}

// Create Role
func (db Database) NewRole(ctx *gin.Context) {

	//init vars
	var role Role
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	//Unmarshal sent json
	if err := ctx.ShouldBindJSON(&role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check field
	if empty_reg.MatchString(role.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid fields"})
		return
	}

	//Create new role instance
	new_role := Role{
		Name: role.Name,
	}

	//create new role
	//check if the role created successfully
	if err := NewRole(db.DB, new_role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Role Created successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Role Created successfully"})
}

// Get all Roles
func (db Database) GetAllRoles(ctx *gin.Context) {

	//get all roles from databse
	roles, err := GetAllRoles(db.DB)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Get all roles
	ctx.JSON(http.StatusOK, roles)
}

// Update Role
func (db Database) UpdateRole(ctx *gin.Context) {

	//init vars
	var role Role
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	//Unmarshal sent json
	if err := ctx.ShouldBindJSON(&role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check field
	if empty_reg.MatchString(role.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid fields"})
		return
	}

	//get the role id
	role_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	//session information
	//extract the user id from the session

	//Create new role instance
	updated_role := Role{
		ID:   uint(role_id),
		Name: role.Name,
	}

	//Update the role
	//Check the role updated successfully
	if err = UpdateRole(db.DB, updated_role); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Role updated successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// Delete the Role
func (db Database) DeleteRole(ctx *gin.Context) {

	//get the role id from the request
	// get id value from path
	role_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check the role deleted successfully
	if err = DeleteRole(db.DB, uint(role_id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Role Deleted successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})

}
