package permission

import (
	"github.com/ezzddinne/CMC_BACKEND/api/app/role"
	"gorm.io/gorm"
)

type CasbinRule struct {
	ID uint   `gorm:"column:id" json:"id"`
	V0 string `gorm:"column:role" json:"role"`
	V1 string `gorm:"column:object" json:"object"`
	V2 string `gorm:"column:action" json:"action"`
}

// Get all permissions
func GetAllPermissions(db *gorm.DB) (permissions []CasbinRule, err error) {
	return permissions, db.Table("casbin_rule").Find(&permissions, "ptype = ?", "p").Error
}

// Get premission by id
func GetPermissionByID(db *gorm.DB, id uint) (permission CasbinRule, err error) {
	return permission, db.Table("casbin_rule").Where("id = ? AND ptype ='p'", id).Find(&permission).Error
}

// Check role exist in permissions
func CheckRoleInPermissions(db *gorm.DB, role_name string) (role role.Role, err error) {
	return role, db.Table("casbin_rule").Where("role = ?", role_name).Find(&role).Error
}

// Check role exist
func CheckRoleExists(db *gorm.DB, name string) (role role.Role, err error) {
	return role, db.Table("roles").Where("name = ?", name).Find(&role).Error
}
