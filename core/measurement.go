package core

type MeasurementUnit string

const (
	PerItem   MeasurementUnit = "item"
	Kilogram  MeasurementUnit = "kilogram"
	Gram      MeasurementUnit = "gram"
	Liter     MeasurementUnit = "liter"
	Mililiter MeasurementUnit = "mililiter"
)

func (m MeasurementUnit) String() string {
	switch m {
	case Kilogram:
		return "KG"
	case Gram:
		return "G"
	case Liter:
		return "L"
	case Mililiter:
		return "ML"
	default:
		return "unknown"
	}
}
