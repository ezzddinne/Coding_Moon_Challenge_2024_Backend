package role

import (
	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RoutesRoles(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// create role route
	router.POST("/new", middleware.Authorize("roles", "write", enforcer), baseInstance.NewRole)

	// get all roles routes
	router.GET("/all", middleware.Authorize("roles", "read", enforcer), baseInstance.GetAllRoles)

	// update role route
	router.PUT("/:id", middleware.Authorize("roles", "write", enforcer), baseInstance.UpdateRole)

	// delete role route
	router.DELETE("/:id", middleware.Authorize("roles", "write", enforcer), baseInstance.DeleteRole)

}
