package domain

import "invest-tracker/internal/i18n"

type AssetType string

const (
	Accion      AssetType = "accion"
	Indice      AssetType = "indice"
	CopyTrading AssetType = "copy_trading"
	Fondo       AssetType = "fondo"
)

func (t AssetType) Valid() bool {
	switch t {
	case Accion, Indice, CopyTrading, Fondo:
		return true
	}
	return false
}

// assetTypeNames son os nomes dos tipos en galego, que son tamén as claves das
// súas traducións.
var assetTypeNames = map[AssetType]string{
	Accion:      "Acción",
	Indice:      "Índice",
	CopyTrading: "Copy-trading",
	Fondo:       "Fondo",
}

// Display devolve o nome do tipo no idioma global (o da CLI).
func (t AssetType) Display() string {
	return t.Label(i18n.Current())
}

// Label devolve o nome do tipo no idioma indicado (p.e. o de cada petición
// web).
func (t AssetType) Label(lang i18n.Lang) string {
	if name, ok := assetTypeNames[t]; ok {
		return lang.T(name)
	}
	return string(t)
}

type Asset struct {
	ID        int64     `json:"id"`
	Type      AssetType `json:"type"`
	Name      string    `json:"name"`
	AmountUSD float64   `json:"amountUsd"`
	Month     int       `json:"month"`
	Year      int       `json:"year"`
}

// Start devolve o mes no que se creou o activo.
func (a Asset) Start() YearMonth {
	return YearMonth{Year: a.Year, Month: a.Month}
}

// CreatedBy indica se o activo xa existía no mes indicado.
func (a Asset) CreatedBy(ym YearMonth) bool {
	return !ym.Before(a.Start())
}

// AssetTypes devolve os tipos de activo nunha orde estable.
func AssetTypes() []AssetType {
	return []AssetType{Accion, Indice, CopyTrading, Fondo}
}
