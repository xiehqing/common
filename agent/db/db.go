package db

import "gorm.io/gorm"

type Queries struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Queries {
	Init(db)
	return &Queries{db: db}
}

func Init(db *gorm.DB) {
	db.AutoMigrate(File{}, Message{}, Session{}, Provider{}, BigModel{})
}
