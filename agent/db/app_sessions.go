package db

import (
	"github.com/pkg/errors"
	"github.com/xiehqing/common/pkg/ormx"
	"gorm.io/gorm"
)

type AppSessions struct {
	ormx.BaseModel
	AppID     int64  `json:"app_id" gorm:"column:app_id;type:bigint;not null"`
	UserID    int64  `json:"user_id" gorm:"column:user_id;type:bigint;not null"`
	SessionID string `json:"session_id" gorm:"column:session_id;type:varchar(255);not null"`
	DataType  string `json:"data_type" gorm:"column:data_type;type:varchar(255);"`
	DataID    string `json:"data_id" gorm:"column:data_id;type:varchar(500);"`
	TenantID  int64  `json:"tenant_id" gorm:"column:tenant_id;type:bigint;"`
}

func (as *AppSessions) TableName() string {
	return "app_sessions"
}

// GetAppSessions 获取应用sessions
func GetAppSessions(db *gorm.DB, appId int64) ([]*AppSessions, error) {
	var appSessions []*AppSessions
	err := db.Where("app_id = ?", appId).Find(&appSessions).Error
	return appSessions, err
}

// CreateAppSession 创建应用session
func CreateAppSession(db *gorm.DB, userId, appId int64, sessionId string) error {
	return db.Where("app_id = ? and session_id = ? and user_id = ?", appId, sessionId, userId).FirstOrCreate(&AppSessions{
		AppID:     appId,
		SessionID: sessionId,
		UserID:    userId,
	}).Error
}

// DeleteAppSession 删除应用session
func DeleteAppSession(db *gorm.DB, sessionId string) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("session_id = ?", sessionId).Delete(&AppSessions{}).Error
		if err != nil {
			return errors.WithMessage(err, "delete app_sessions error")
		}
		return nil
	})
	return err
}

// GetDataAppSession 获取应用数据session
func GetDataAppSession(db *gorm.DB, tenantId, userId, appId int64, dataType, dataId string) (*AppSessions, error) {
	var appSession *AppSessions
	err := db.Where("app_id = ? and user_id = ? and tenant_id = ? and data_type = ? and data_id = ?",
		appId, userId, tenantId, dataType, dataId).First(&appSession).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return appSession, nil
}

// CreateDataAppSession 创建应用数据session
func CreateDataAppSession(db *gorm.DB, tenantId, userId, appId int64, dataType, dataId, sessionId string) error {
	return db.Where("app_id = ? and session_id = ? and user_id = ? and data_type = ? and data_id = ? and tenant_id = ?",
		appId, sessionId, userId, dataType, dataId, tenantId).FirstOrCreate(&AppSessions{
		AppID:     appId,
		SessionID: sessionId,
		UserID:    userId,
		DataID:    dataId,
		DataType:  dataType,
		TenantID:  tenantId,
	}).Error
}

// GetAppSessionsByDataIds 根据dataId获取Agent会话
func GetAppSessionsByDataIds(db *gorm.DB, appId int64, dataType string, dataId []string) ([]*AppSessions, error) {
	var agentConversations []*AppSessions
	err := db.Where("app_id = ? and data_type = ? and data_id in ?", appId, dataType, dataId).Find(&agentConversations).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "查询APP会话失败")
	}
	return agentConversations, nil
}
