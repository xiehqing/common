package db

import (
	"context"
)

// GetProviders 获取所有provider
//func (q *Queries) GetProviders(ctx context.Context) ([]catwalk.Provider, error) {
//	var providers []*Provider
//	err := q.db.Find(&providers).Error
//	if err != nil {
//		return nil, err
//	}
//	var models []*BigModel
//	err = q.db.Find(&models).Error
//	if err != nil {
//		return nil, err
//	}
//	modeMap := make(map[string][]*BigModel)
//	for _, model := range models {
//		modeMap[model.ProviderID] = append(modeMap[model.ProviderID], model)
//	}
//
//	var result []catwalk.Provider
//	for _, provider := range providers {
//		var catwalkModels []catwalk.Model
//		var defaultLargeModelID string
//		var defaultSmallModelID string
//		if dbModels, ok := modeMap[provider.ID]; ok {
//			for _, dbModel := range dbModels {
//				if dbModel.IsDefaultBigModel {
//					defaultLargeModelID = dbModel.ID
//				}
//				if dbModel.IsDefaultSmallModel {
//					defaultSmallModelID = dbModel.ID
//				}
//				catwalkModels = append(catwalkModels, catwalk.Model{
//					ID:                 dbModel.ID,
//					Name:               dbModel.Name,
//					CostPer1MIn:        dbModel.CostPer1mIn,
//					CostPer1MOut:       dbModel.CostPer1mOut,
//					CostPer1MInCached:  dbModel.CostPer1mInCached,
//					CostPer1MOutCached: dbModel.CostPer1mOutCached,
//					ContextWindow:      dbModel.ContextWindow,
//					CanReason:          dbModel.CanReason,
//					DefaultMaxTokens:   dbModel.DefaultMaxTokens,
//				})
//			}
//		}
//		if len(catwalkModels) > 0 {
//			var headers map[string]string
//			if provider.DefaultHeaders != "" {
//				json.Unmarshal([]byte(provider.DefaultHeaders), &headers)
//			}
//			if defaultLargeModelID == "" {
//				defaultLargeModelID = catwalkModels[0].ID
//			}
//			if defaultSmallModelID == "" {
//				defaultSmallModelID = catwalkModels[0].ID
//			}
//			result = append(result, catwalk.Provider{
//				Name:                provider.Name,
//				ID:                  catwalk.InferenceProvider(provider.ID),
//				APIKey:              provider.ApiKey,
//				APIEndpoint:         provider.ApiEndpoint,
//				Type:                catwalk.Type(provider.Type),
//				DefaultLargeModelID: defaultLargeModelID,
//				DefaultSmallModelID: defaultSmallModelID,
//				DefaultHeaders:      headers,
//				Models:              catwalkModels,
//			})
//		}
//	}
//	return result, nil
//}

func (q *Queries) GetProviders(ctx context.Context) ([]Provider, error) {
	var providers []Provider
	err := q.db.Find(&providers).Error
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (q *Queries) GetBigModels(ctx context.Context) ([]BigModel, error) {
	var bigModels []BigModel
	err := q.db.Find(&bigModels).Error
	if err != nil {
		return nil, err
	}
	return bigModels, nil
}
