package repositories

import (
	"project/database"
	"project/models/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetAllRoot() *gorm.DB {
	return database.DB.Db.
		Model(&entity.Project{})
}

func GetAllByCompany(companyID string) *gorm.DB {
	return database.DB.Db.
		Model(&entity.Project{}).
		Where("company_id = ?", companyID)
}

func GetProjectListRegularUser(companyID, username string, projectIDs []uuid.UUID) *gorm.DB {
	return database.DB.Db.
		Model(&entity.Project{}).
		Where("company_id = ?", companyID).
		Where("id IN (?)", projectIDs)
}
