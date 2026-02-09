package provider

import (
	"context"
	"encoding/json"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/xiehqing/common/agent/db"
)

type Service interface {
	List(ctx context.Context) ([]catwalk.Provider, error)
}

type service struct {
	q db.Querier
}

func NewService(q db.Querier) Service {
	return &service{q: q}
}

func (s *service) List(ctx context.Context) ([]catwalk.Provider, error) {
	providers, err := s.q.GetProviders(ctx)
	if err != nil {
		return nil, err
	}
	models, err := s.q.GetBigModels(ctx)
	if err != nil {
		return nil, err
	}
	modeMap := make(map[string][]db.BigModel)
	for _, model := range models {
		modeMap[model.ProviderID] = append(modeMap[model.ProviderID], model)
	}

	var result []catwalk.Provider
	for _, provider := range providers {
		var catwalkModels []catwalk.Model
		var defaultLargeModelID string
		var defaultSmallModelID string
		if dbModels, ok := modeMap[provider.ID]; ok {
			for _, dbModel := range dbModels {
				if dbModel.IsDefaultBigModel {
					defaultLargeModelID = dbModel.ID
				}
				if dbModel.IsDefaultSmallModel {
					defaultSmallModelID = dbModel.ID
				}
				catwalkModels = append(catwalkModels, catwalk.Model{
					ID:                 dbModel.ID,
					Name:               dbModel.Name,
					CostPer1MIn:        dbModel.CostPer1mIn,
					CostPer1MOut:       dbModel.CostPer1mOut,
					CostPer1MInCached:  dbModel.CostPer1mInCached,
					CostPer1MOutCached: dbModel.CostPer1mOutCached,
					ContextWindow:      dbModel.ContextWindow,
					CanReason:          dbModel.CanReason,
					DefaultMaxTokens:   dbModel.DefaultMaxTokens,
				})
			}
		}
		if len(catwalkModels) > 0 {
			var headers map[string]string
			if provider.DefaultHeaders != "" {
				json.Unmarshal([]byte(provider.DefaultHeaders), &headers)
			}
			if defaultLargeModelID == "" {
				defaultLargeModelID = catwalkModels[0].ID
			}
			if defaultSmallModelID == "" {
				defaultSmallModelID = catwalkModels[0].ID
			}
			result = append(result, catwalk.Provider{
				Name:                provider.Name,
				ID:                  catwalk.InferenceProvider(provider.ID),
				APIKey:              provider.ApiKey,
				APIEndpoint:         provider.ApiEndpoint,
				Type:                catwalk.Type(provider.Type),
				DefaultLargeModelID: defaultLargeModelID,
				DefaultSmallModelID: defaultSmallModelID,
				DefaultHeaders:      headers,
				Models:              catwalkModels,
			})
		}
	}
	return result, nil
}
