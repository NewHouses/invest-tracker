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
	ID        int64
	Type      AssetType
	Name      string
	AmountUSD float64
	Month     int
	Year      int
}
