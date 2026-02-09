package db

import (
	"github.com/xiehqing/common/models"
	"github.com/xiehqing/common/pkg/ormx"
	"gorm.io/gorm"
)

type AppUploadFiles struct {
	ormx.BaseModel
	AppID      int64  `json:"app_id" gorm:"column:app_id;type:bigint;not null"`
	Name       string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Size       int64  `json:"size" gorm:"column:size;type:bigint;not null"`
	OriginPath string `json:"originPath" gorm:"column:origin_path;type:varchar(255);not null"`
	TargetPath string `json:"targetPath" gorm:"column:target_path;type:varchar(255);not null"`
	UserID     int64  `json:"user_id" gorm:"column:user_id;type:bigint;not null"`
	SessionID  string `json:"session_id" gorm:"column:session_id;type:varchar(255);not null"`
}

func (af *AppUploadFiles) TableName() string {
	return "app_upload_files"
}

// CreateAppUploadFiles 保存应用上传文件
func CreateAppUploadFiles(db *gorm.DB, files []*AppUploadFiles) error {
	return models.CreateInBatches(db, files)
}

// UpdateAppUploadFiles 更新应用上传文件
func UpdateAppUploadFiles(db *gorm.DB, files []*AppUploadFiles) error {
	return models.Update(db, files)
}

// GetAppUploadFiles 获取应用上传文件
func GetAppUploadFiles(db *gorm.DB, appID int64, userID *int64, sessionID string) ([]*AppUploadFiles, error) {
	var files []*AppUploadFiles
	tx := db.Model(&AppUploadFiles{}).Where("app_id = ?", appID)
	if userID != nil {
		tx = tx.Where("user_id = ?", *userID)
	}
	if sessionID != "" {
		tx = tx.Where("session_id = ?", sessionID)
	}
	err := tx.Find(&files).Error
	return files, err
}

// GetAppSessionUploadFiles 获取应用上传文件
func GetAppSessionUploadFiles(db *gorm.DB, sessionID string, targetFiles []string) ([]*AppUploadFiles, error) {
	var files []*AppUploadFiles
	tx := db.Model(&AppUploadFiles{}).Where("session_id = ?", sessionID)
	if len(targetFiles) > 0 {
		tx = tx.Where("target_path IN ?", targetFiles)
	}
	err := tx.Find(&files).Error
	return files, err
}
