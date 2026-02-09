package db

import (
	"context"
	"github.com/xiehqing/common/pkg/ormx"
)

type Apps struct {
	ormx.BaseModel
	Name         string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Code         string `json:"code" gorm:"column:code;type:varchar(255);not null"`
	Icon         string `json:"icon" gorm:"column:icon;type:varchar(255);not null"`
	Description  string `json:"description" gorm:"column:description;type:varchar(2000);"`
	Tags         string `json:"tags" gorm:"column:tags;type:varchar(500);"`
	Prompt       string `json:"prompt" gorm:"column:prompt;type:varchar(2000);"`
	QuickPrompts string `json:"quickPrompts" gorm:"column:quick_prompts;type:text;"`
	WorkingDir   string `json:"workingDir" gorm:"column:working_dir;type:varchar(500);not null"`
	DataDir      string `json:"dataDir" gorm:"column:data_dir;type:varchar(500);not null"`
	Status       int    `json:"status" gorm:"column:status;type:int(11);not null"`
	Background   string `json:"background" gorm:"column:background;type:varchar(500);"`
	ProviderID   string `json:"providerId" gorm:"column:provider_id;type:varchar(255);"`
	BigModelID   string `json:"bigModelId" gorm:"column:big_model_id;type:varchar(255);"`
}

func (a *Apps) TableName() string {
	return "apps"
}

// GetAllApps 获取所有应用
func (q *Queries) GetAllApps(ctx context.Context, status *int) ([]*Apps, error) {
	var apps []*Apps
	if status != nil {
		q.db = q.db.Where("status = ?", *status)
	}
	err := q.db.Find(&apps).Error
	return apps, err
}

// UpdateApp 更新应用
func (q *Queries) UpdateApp(ctx context.Context, app *Apps) error {
	return q.db.Save(app).Error
}
