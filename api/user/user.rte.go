package user

import (
	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RoutesAuth(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// Create leader route
	router.POST("/new", baseInstance.NewLeader)

	//verify email
	router.POST("/verify/:email", baseInstance.handleEmailVerification)

	// Sign in to squad account route
	router.POST("/signin", baseInstance.SignInLeader)

	//upload cv leader route
	router.POST("/file/:email", baseInstance.LeaderFileUpload())

}

func RouteAdmin(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// Change paiment status route
	router.PATCH("/:id", middleware.Authorize("paiment", "write", enforcer), baseInstance.ChangePaimentStatus)
}

func RoutesUsersJWT(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// Get all users route
	router.GET("/allusers", middleware.Authorize("users", "read", enforcer), baseInstance.GetAllUsers)

	// Get user by id route
	router.GET("/:id", middleware.Authorize("users", "write", enforcer), baseInstance.GetUserByID)

	// Get user by id to merge squad
	router.GET("/id", middleware.Authorize("users", "write", enforcer), baseInstance.GetUserByIDFront)

	// Get users by role route
	router.GET("/role/:role", middleware.Authorize("users", "read", enforcer), baseInstance.GetUsersByRole)

	// Get users by squad id
	router.GET("/squad/:id", middleware.Authorize("users", "read", enforcer), baseInstance.GetUsersBySquadID)

	// Delete user route
	router.DELETE("/:id", middleware.Authorize("users", "write", enforcer), baseInstance.DeleteUser)


}

func RoutesUserPassword(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	//Forget Password route
	router.POST("/forgotpassword", baseInstance.ForgetPassword)

	//Reset Password route
	router.PATCH("/resetpassword", baseInstance.ResetPassword)

}
