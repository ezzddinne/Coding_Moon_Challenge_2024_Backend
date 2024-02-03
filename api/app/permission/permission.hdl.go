package permission

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

// create new permission
func (db Database) NewPermission(ctx *gin.Context) {

	//init vars
	var permission CasbinRule
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	// check if the sent content is compatible with CasbinRule
	if err := ctx.ShouldBindJSON(&permission); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//unmarshall sent json
	// check fields
	if empty_reg.MatchString(permission.V0) || empty_reg.MatchString(permission.V1) || empty_reg.MatchString(permission.V2) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid fields"})
		return
	}

	//check the action passed by the user
	if permission.V2 != "read" && permission.V2 != "write" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "permission is invalid"})
		return
	}

	role, err := CheckRoleExists(db.DB, permission.V0)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//check the role name is not empty
	if role.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "role name is invalid"})
		return
	}

	//Check the role has the policy or not
	//if not add a policy to it
	if hasPolicy := db.Enforcer.HasPolicy(permission.V0, permission.V1, permission.V2); !hasPolicy {
		db.Enforcer.AddPolicy(permission.V0, permission.V1, permission.V2)
	}

	//permission created successfully
	ctx.JSON(http.StatusOK, gin.H{"message": "Permission Created successfuly"})

}

// Get all permissions
func (db Database) GetAllPermissions(ctx *gin.Context) {

	//Get the permessions
	permissions, err := GetAllPermissions(db.DB)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Return the permissions
	ctx.JSON(http.StatusOK, permissions)
}

// Get permission by id
func (db Database) GetPermissionByID(ctx *gin.Context) {

	//Get permission id from the request path
	permission_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Get th permission by the id
	permission, err := GetPermissionByID(db.DB, uint(permission_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//check if the permission is not completed
	if permission.V0 == "" || permission.V1 == "" || permission.V2 == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please complete all fields"})
		return
	}

	//Return the permission object
	ctx.JSON(http.StatusOK, permission)

}

// Update permission
func (db Database) UpdatePermission(ctx *gin.Context) {

	//init vars
	var permission CasbinRule
	empty_reg, _ := regexp.Compile(os.Getenv("EMPTY_REGEX"))

	//unmarshall sent json
	// check fields
	if empty_reg.MatchString(permission.V0) || empty_reg.MatchString(permission.V1) || empty_reg.MatchString(permission.V2) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid fields"})
		return
	}

	//check role exists
	role, err := CheckRoleExists(db.DB, permission.V0)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//check the action passed by the user
	if permission.V2 != "read" && permission.V2 != "write" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "permission is invalid"})
		return
	}

	// check if role name is empty
	if role.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "role name is invalid"})
		return
	}

	// get the id from the query
	permission_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// get the row of the query
	db_permission, err := GetPermissionByID(db.DB, uint(permission_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check the permission it's a policy role
	// check if it is a policy role
	if permission.V0 == "" || permission.V1 == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "permission not found"})
		return
	}

	//update the policy
	//check if they have the same object
	if permission.V1 == db_permission.V1 {
		db.Enforcer.UpdatePolicy([]string{db_permission.V0, db_permission.V1, db_permission.V2}, []string{permission.V0, permission.V1, permission.V2})
		ctx.JSON(http.StatusOK, gin.H{"message": "permission updated successfully"})
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid edit"})
	}
}

// Delete permission
func (db Database) DeletePermission(ctx *gin.Context) {

	// get id value from path
	permission_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// get the row of the query
	permission, err := GetPermissionByID(db.DB, uint(permission_id))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check if it is a policy role
	if permission.V0 == "" || permission.V1 == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "permission not found"})
		return
	}

	// delete the row in the table
	_, err = db.Enforcer.RemovePolicy(permission.V0, permission.V1, permission.V2)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "permission removed successfully"})
}
