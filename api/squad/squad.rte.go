package squad

import (
	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RoutesAuthJWT(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// create squad route
	router.POST("/new", middleware.Authorize("squads", "write", enforcer), baseInstance.CreateSquad)

	// Get all squads route
	router.GET("/allsquads", middleware.Authorize("squads", "read", enforcer), baseInstance.GetAllSquads)

	// Get squad by id
	router.GET("/squad/:id", middleware.Authorize("squads", "read", enforcer), baseInstance.GetSquadByID)

	// get squad id front
	//Front
	router.GET("/squad/id", middleware.Authorize("squads", "write", enforcer), baseInstance.GetSquadByIDFront)

	// get logo
	//front
	router.GET("/squad/logo", middleware.Authorize("squads", "write", enforcer), baseInstance.GetSquadLogo)

	// Get squad by email
	router.GET("/:email", middleware.Authorize("squads", "read", enforcer), baseInstance.GetSquadByEmail)

	// Delete squad route
	router.DELETE("/delete", middleware.Authorize("squads", "write", enforcer), baseInstance.DeleteSquad)

	// add member route
	router.POST("/add", middleware.Authorize("squads", "write", enforcer), baseInstance.AddMember)

	// upload image route
	router.POST("/image", middleware.Authorize("squads", "write", enforcer), baseInstance.ImageUpload())

	// update logo route
	router.PATCH("/update/logo", middleware.Authorize("squads", "write", enforcer), baseInstance.ImageUpload())

	// upload file route
	router.POST("/file", middleware.Authorize("squads", "write", enforcer), baseInstance.FileUpload())

	// update squad name route
	router.PATCH("/update", middleware.Authorize("squads", "write", enforcer), baseInstance.UpdateName)

	// Delete leader route
	router.DELETE("/leader", middleware.Authorize("squads", "write", enforcer), baseInstance.DeleteLeader)

}
