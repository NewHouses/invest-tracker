package domain

type InvestmentType string

const (
	Accion      InvestmentType = "accion"
	Indice      InvestmentType = "indice"
	CopyTrading InvestmentType = "copy_trading"
	Fondo       InvestmentType = "fondo"
)

func (t InvestmentType) Valid() bool {
	switch t {
	case Accion, Indice, CopyTrading, Fondo:
		return true
	}
	return false
}

func (t InvestmentType) Display() string {
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

type Investment struct {
	ID        int64
	Type      InvestmentType
	Name      string
	AmountUSD float64
	Month     int
	Year      int
}
