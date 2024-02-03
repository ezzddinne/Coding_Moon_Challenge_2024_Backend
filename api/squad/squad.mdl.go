package squad

import (
	"mime/multipart"

	"github.com/ezzddinne/CMC_BACKEND/api/user"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Squad struct {
	ID           uint           `gorm:"column:id;autoIncrement;primaryKey" json:"id"`
	Name         string         `gorm:"column:name;not null;unique" json:"name"`
	CreatedBy    uint           `gorm:"column:created_by;not null" json:"created_by"`
	LeaderID     user.User      `gorm:"foreignKey:CreatedBy;references:ID"`
	SquadMembers pq.Int32Array  `gorm:"column:squad_members;type:integer[]" json:"squad_members"`
	LogoURL      string         `gorm:"column:logo_url;not null" json:"logo_url"`
	CvURLS       pq.StringArray `gorm:"column:cv_urls;type:varchar[]" json:"cv_urls"`

	gorm.Model
}

type File struct {
	File multipart.File `json:"file,omitempty" validate:"required"`
}

// create new squad
func NewSquad(db *gorm.DB, squad Squad) (Squad, error) {
	return squad, db.Create(&squad).Error
}

// get all users
func GetAllSquads(db *gorm.DB) (squads []Squad, err error) {
	return squads, db.Preload("LeaderID").Find(&squads).Error
}

// update function
func UpdateSquad(db *gorm.DB, squad Squad) error {
	return db.Where("id = ?", squad.ID).Updates(&squad).Error
}

// delete squad
func DeleteSquad(db *gorm.DB, squad_id uint) error {
	return db.Where("id = ?", squad_id).Delete(&Squad{}).Error
}

// get squad by email
func GetSquadByEmail(db *gorm.DB, email string) (squad Squad, err error) {
	return squad, db.Preload("LeaderID").First(&squad, "email=?", email).Error
}

// Get squad by id
func GetSquadByID(db *gorm.DB, squad_id uint) (squad Squad, err error) {
	return squad, db.Where("id = ?", squad_id).Preload("LeaderID").First(&squad).Error
}

// Check user already created a squad
// check user existence
func CheckUserCreateSquad(db *gorm.DB, id uint) bool {

	//init vars
	user := &Squad{}

	//check if user exist
	check := db.Where("created_by = ?", id).First(user)
	if check.Error != nil {
		return false
	}

	if check.RowsAffected == 0 {
		return false
	} else {
		return true
	}
}
