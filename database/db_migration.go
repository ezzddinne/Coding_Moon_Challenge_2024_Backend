package database

import (
	"fmt"
	"os"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/ezzddinne/CMC_BACKEND/api/app/permission"
	"github.com/ezzddinne/CMC_BACKEND/api/app/role"
	"github.com/ezzddinne/CMC_BACKEND/api/squad"
	"github.com/ezzddinne/CMC_BACKEND/api/user"
	"gorm.io/gorm"
)

// auto migrate dtables
func _auto_migrate_tables(db *gorm.DB) {

	// auto migrate the casbin table
	if err := db.Table("casbin_rule").AutoMigrate(&permission.CasbinRule{}); err != nil {
		panic(fmt.Sprintf("Error while creating the casbin table : %v", err))
	}

	// auto migrate user, role & squad tables
	if err := db.AutoMigrate(
		&role.Role{},
		&squad.Squad{},
		&user.User{},
	); err != nil {
		panic(err)
	}

}

// auto create root user
func _create_root_user(db *gorm.DB, enforcer *casbin.Enforcer) {

	//init vars
	//root
	var user_id uint
	root_role := role.Role{}
	root_user := user.User{}
	root_squad := squad.Squad{}
	//default role ==> user
	user_role := role.Role{}

	//create root role ==> root
	//check root role exists
	if check := db.Where("name = ?", os.Getenv("DEFAULT_ROOT")).Find(&root_role); check.RowsAffected == 0 && check.Error == nil {

		//create root role
		db_role := role.Role{Name: os.Getenv("DEFAULT_ROOT")}

		if err := db.Create(&db_role).Error; err != nil {
			panic(fmt.Sprintf("[WARNING] error while creating the root role: %v", err))
		}
	}

	//create root user ==> Leader
	//check root user exists
	if check := db.Where("email = ?", os.Getenv("DEFAULT_EMAIL")).Find(&root_user); check.RowsAffected == 0 && check.Error == nil {

		//convert the paiment status to bool
		defaultPaymentStatus := os.Getenv("DEFAULT_PAIMENT_STATUS")
		paymentStatus := defaultPaymentStatus == "true"

		//create the root user
		db_user := user.User{FirstName: os.Getenv("DEFAULT_FIRSTNAME"), LastName: os.Getenv("DEFAULT_LASTNAME"), Email: os.Getenv("DEFAULT_EMAIL"), University: os.Getenv("DEFAULT_UNIVERSITY"), Password: os.Getenv("DEFAULT_USER_PASSWORD"), Phone: os.Getenv("DEFAULT_PHONE"), Paiment_Status: paymentStatus, Role: os.Getenv("DEFAULT_ROOT")}
		user.HashPassword(&db_user.Password)

		if err := db.Create(&db_user).Error; err != nil {
			panic(fmt.Sprintf("[WARNING] error while creating the root user: %v", err))
		}

		// used to save user id to create squad if the root user dosen't exist yet
		user_id = db_user.ID
	} else {

		// used to save user id to create squad if the root user exists
		user_id = root_user.ID

	}

	// add policy
	enforcer.AddGroupingPolicy(strconv.FormatUint(uint64(user_id), 10), os.Getenv("DEFAULT_ROOT"))

	// create default user ==> member
	if check := db.Where("name = ?", os.Getenv("DEFAULT_USER")).Find(&user_role); check.RowsAffected == 0 && check.Error == nil {

		// create role user
		db_role := role.Role{Name: os.Getenv("DEFAULT_USER")}

		if err := db.Create(&db_role).Error; err != nil {
			panic(fmt.Sprintf("[WARNING] error while creating the user role: %v", err))
		}
	}

	// add policy
	enforcer.AddGroupingPolicy(strconv.FormatUint(uint64(0), 10), os.Getenv("DEFAULT_USER"))

	// create squad
	//check squad exists
	if check := db.Where("name = ?", os.Getenv("DEFAULT_SQUAD_NAME")).Find(&root_squad); check.RowsAffected == 0 && check.Error == nil {

		//init vars
		var intIDs []int32
		intIDs = append(intIDs, int32(user_id))

		//create sqaud
		db_squad := &squad.Squad{Name: os.Getenv("DEFAULT_SQUAD_NAME"), CreatedBy: user_id, SquadMembers: intIDs}

		err := db.Create(&db_squad).Error
		if err != nil {
			panic(fmt.Sprintf("[WARNING] error while creating the root squad: %v", err))
		}

		//edit user to add squad id
		if check := db.Where("email = ?", os.Getenv("DEFAULT_EMAIL")).Find(&root_user); check.RowsAffected == 1 && check.Error == nil {
			root_user.SquadID = db_squad.ID
			if update := db.Where("id = ?", root_user.ID).Updates(&root_user); update.Error != nil {
				panic(fmt.Sprintf("[WARNING] error while updating the root user: %v", update.Error))
			}
		}

	}
}

func AutoMigrateDatabase(db *gorm.DB, enforcer *casbin.Enforcer) {

	// create tables
	_auto_migrate_tables(db)

	//create root
	_create_root_user(db, enforcer)
}
