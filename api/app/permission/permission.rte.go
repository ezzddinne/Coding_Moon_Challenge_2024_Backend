package permission

import (
	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RoutesPermissions(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	baseInstance := Database{DB: db, Enforcer: enforcer}

	// create permission route
	router.POST("/new", middleware.Authorize("permissions", "write", enforcer), baseInstance.NewPermission)

	// get all permission route
	router.GET("/all", middleware.Authorize("permissions", "read", enforcer), baseInstance.GetAllPermissions)

	// get permission by id route
	router.GET("/:id", middleware.Authorize("permissions", "read", enforcer), baseInstance.GetPermissionByID)

	// update permission route
	router.PUT("/:id", middleware.Authorize("permissions", "write", enforcer), baseInstance.UpdatePermission)

	// delete permission route
	router.DELETE("/:id", middleware.Authorize("permissions", "write", enforcer), baseInstance.DeletePermission)
}
