package role

import (
	"gorm.io/gorm"
) 

type Role struct {
	ID        uint   `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	Name      string `gorm:"column:name;not null;unique" json:"name"`
	gorm.Model
}

//Create new role
func NewRole(db *gorm.DB, role Role) error {
	return db.Create(&role).Error
}

//Get all roles
func GetAllRoles(db *gorm.DB) (roles []Role, err error) {
	return roles, db.Find(&roles).Error
}

//Get Role By name
func GetRoleByName(db *gorm.DB, name string) (role Role, err error) {
	return role, db.Where("name = ?",name).First(&role).Error
}

//Update role
func UpdateRole(db *gorm.DB, role Role) error {
	return db.Where("role_id = ?",role.ID).Updates(&role).Error
}

//Delete role
func DeleteRole(db *gorm.DB, role_id uint) error {
	return db.Where("role_id = ?",role_id).Delete(&Role{}).Error
}

