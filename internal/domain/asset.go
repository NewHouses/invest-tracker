package domain

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

func (t AssetType) Display() string {
	switch t {
	case Accion:
		return "Acción"
	case Indice:
		return "Índice"
	case CopyTrading:
		return "Copy-trading"
	case Fondo:
		return "Fondo"
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
