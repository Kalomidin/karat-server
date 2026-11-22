package domain

// Error is a client-visible failure. Code is stable for the app to translate.
type Error struct {
	Code string
}

func (e *Error) Error() string { return e.Code }

func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && other.Code == e.Code
}

var (
	ErrInvalidJSON        = &Error{Code: "invalid_json"}
	ErrInvalidName        = &Error{Code: "invalid_name"}
	ErrInvalidStockNumber = &Error{Code: "invalid_stock_number"}
	ErrInvalidKind        = &Error{Code: "invalid_kind"}
	ErrInvalidMetal       = &Error{Code: "invalid_metal"}
	ErrInvalidPurity      = &Error{Code: "invalid_purity"}
	ErrInvalidWeight      = &Error{Code: "invalid_weight"}
	ErrInvalidCost        = &Error{Code: "invalid_cost"}
	ErrInvalidNotes       = &Error{Code: "invalid_notes"}
	ErrInvalidDate        = &Error{Code: "invalid_date"}
	ErrInvalidPrice       = &Error{Code: "invalid_price"}
	ErrFutureSale         = &Error{Code: "future_sale"}
	ErrAlreadySold        = &Error{Code: "already_sold"}
	ErrStillInStock       = &Error{Code: "still_in_stock"}
	ErrSoldLocked         = &Error{Code: "sold_locked"}
	ErrMissing            = &Error{Code: "missing"}
)
