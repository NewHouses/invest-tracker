package web

import (
	"net/http"

	"invest-tracker/internal/domain"
)

type metaAssetType struct {
	Value domain.AssetType `json:"value"`
	Label string           `json:"label"`
}

type metaResponse struct {
	AssetTypes []metaAssetType `json:"assetTypes"`
	MinYear    int             `json:"minYear"`
	MaxYear    int             `json:"maxYear"`
}

// registerMetaRoutes rexistra as rutas de metadatos da API.
func (s *Server) registerMetaRoutes() {
	s.handleAPI("GET /api/meta", func(w http.ResponseWriter, r *http.Request) {
		types := domain.AssetTypes()
		items := make([]metaAssetType, 0, len(types))
		for _, typ := range types {
			items = append(items, metaAssetType{Value: typ, Label: typ.Display()})
		}
		writeJSON(w, http.StatusOK, metaResponse{AssetTypes: items, MinYear: domain.MinYear, MaxYear: domain.MaxYear})
	})
}
