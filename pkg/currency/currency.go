package currency

import "github.com/google/uuid"

// Stable ids for the three shop currencies. Items store one of these.
var (
	DollarID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	RubleID  = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	SomoniID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
)

type Currency struct {
	ID   uuid.UUID
	Code string
	Name string
}

func Catalog() []Currency {
	return []Currency{
		{ID: DollarID, Code: "USD", Name: "Dollar"},
		{ID: RubleID, Code: "RUB", Name: "Ruble"},
		{ID: SomoniID, Code: "TJS", Name: "Somoni"},
	}
}
