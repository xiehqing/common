package db

import (
	"github.com/xiehqing/common/pkg/ormx"
	"gorm.io/gorm"
)

type File struct {
	ormx.UuidModel
	SessionID string `json:"session_id" gorm:"type:varchar(255);not null;comment:'session_id'"`
	Path      string `json:"path" gorm:"type:varchar(255);not null;comment:'path'"`
	Content   string `json:"content" gorm:"type:longtext;comment:'content'"`
	Version   int64  `json:"version" gorm:"type:int(11);not null;comment:'version'"`
}

func (f *File) TableName() string {
	return "files"
}

type Message struct {
	ormx.UuidModel
	SessionID        string `json:"sessionId" gorm:"type:varchar(255);not null;column:session_id;comment:'session_id'"`
	Role             string `json:"role" gorm:"type:varchar(255);not null;column:role;comment:'role'"`
	Parts            string `json:"parts" gorm:"type:longtext;column:parts;comment:'parts'"`
	Model            string `json:"model" gorm:"type:varchar(255);column:model;comment:'model'"`
	FinishedAt       int64  `json:"finishedAt" gorm:"type:bigint(20);column:finished_at;comment:'结束时间'"`
	Provider         string `json:"provider" gorm:"type:varchar(255);column:provider;comment:'provider'"`
	IsSummaryMessage int64  `json:"isSummaryMessage" gorm:"type:int(11);column:is_summary_message;comment:'是否是summary_message'"`
}

func (m *Message) TableName() string {
	return "messages"
}

type Session struct {
	ormx.UuidModel
	ParentSessionID  string  `json:"parentSessionId" gorm:"type:varchar(255);not null;column:parent_session_id;comment:'parent_session_id'"`
	Title            string  `json:"title" gorm:"type:varchar(255);not null;comment:'title';column:title"`
	MessageCount     int64   `json:"messageCount" gorm:"type:int(11);not null;comment:'message_count';column:message_count"`
	PromptTokens     int64   `json:"promptTokens" gorm:"type:int(11);not null;comment:'prompt_tokens';column:prompt_tokens"`
	CompletionTokens int64   `json:"completionTokens" gorm:"type:int(11);not null;comment:'completion_tokens';column:completion_tokens"`
	Cost             float64 `json:"cost" gorm:"type:decimal(10,2);not null;comment:'cost';column:cost"`
	SummaryMessageID string  `json:"summaryMessageId" gorm:"type:varchar(255);not null;column:summary_message_id;comment:'summary_message_id'"`
	Todos            string  `json:"todos" gorm:"type:text;comment:'todos';column:todos"`
}

func (s *Session) TableName() string {
	return "sessions"
}

type Provider struct {
	ormx.UuidModel
	Name           string `json:"name" gorm:"type:varchar(255);column:name;not null"`
	ApiKey         string `json:"apiKey" gorm:"type:varchar(255);column:api_key;comment:'api_key';not null"`
	ApiEndpoint    string `json:"apiEndpoint" gorm:"type:varchar(255);column:api_endpoint;comment:'api_endpoint';not null"`
	Type           string `json:"type" gorm:"type:varchar(255);not null;column:type"`
	DefaultHeaders string `json:"defaultHeaders" gorm:"type:varchar(255);not null;column:default_headers"`
}

func (p *Provider) TableName() string {
	return "provider"
}

type BigModel struct {
	ormx.UuidModel
	ProviderID          string  `json:"providerId" gorm:"type:varchar(255);not null"`
	Name                string  `json:"name" gorm:"type:varchar(255);not null"`
	CostPer1mIn         float64 `json:"costPer1mIn" gorm:"type:decimal(10,2);comment:'cost_per_1m_in';not null"`
	CostPer1mOut        float64 `json:"costPer1mOut" gorm:"type:decimal(10,2);comment:'cost_per_1m_out';not null"`
	CostPer1mInCached   float64 `json:"costPer1mInCached" gorm:"type:decimal(10,2);comment:'cost_per_1m_in_cached';not null"`
	CostPer1mOutCached  float64 `json:"costPer1mOutCached" gorm:"type:decimal(10,2);comment:'cost_per_1m_out_cached';not null"`
	ContextWindow       int64   `json:"contextWindow" gorm:"type:int(11);comment:'context_window';not null"`
	DefaultMaxTokens    int64   `json:"defaultMaxTokens" gorm:"type:int(11);comment:'default_max_tokens';not null"`
	CanReason           bool    `json:"canReason" gorm:"type:tinyint(1);comment:'can_reason';not null"`
	SupportsAttachments bool    `json:"supportsAttachments" gorm:"type:tinyint(1);comment:'supports_attachments';not null"`
	IsDefaultSmallModel bool    `json:"isDefaultSmallModel" gorm:"type:tinyint(1);comment:'is_default_small_model';not null"`
	IsDefaultBigModel   bool    `json:"isDefaultBigModel" gorm:"type:tinyint(1);comment:'is_default_big_model';not null"`
}

func (b *BigModel) TableName() string {
	return "models"
}

// NewDefaultModel 创建默认model
func NewDefaultModel(ID, name string, canReason bool) *BigModel {
	return &BigModel{
		UuidModel: ormx.UuidModel{
			ID: ID,
		},
		Name:                name,
		CostPer1mIn:         3,
		CostPer1mOut:        15,
		CostPer1mInCached:   3.75,
		CostPer1mOutCached:  0.3,
		CanReason:           canReason,
		ContextWindow:       200000,
		DefaultMaxTokens:    50000,
		SupportsAttachments: true,
	}
}

// GetBigModelByID 根据ID获取model
func GetBigModelByID(db *gorm.DB, id string) (*Provider, *BigModel, error) {
	var model *BigModel
	var provider *Provider
	err := db.Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, nil, err
	}
	err = db.Where("id = ?", model.ProviderID).First(&provider).Error
	if err != nil {
		return nil, nil, err
	}
	return provider, model, nil
}
