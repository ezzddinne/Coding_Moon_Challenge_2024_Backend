package app

import (
	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/api/app/permission"
	"github.com/ezzddinne/CMC_BACKEND/api/app/role"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// declare app routes
func RoutesApps(router *gin.RouterGroup, db *gorm.DB, enforcer *casbin.Enforcer) {

	// role routes
	role.RoutesRoles(router.Group("/role"), db, enforcer)

	// permission routes
	permission.RoutesPermissions(router.Group("/permission"), db, enforcer)

}
