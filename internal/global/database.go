package global

import "gorm.io/gorm"

func SetDatabase(db *gorm.DB) {
	DB = db
}
