package common

// Error is a failure the client can read. Code stays stable for the app.
type Error struct {
	code   string
	status int
}

func NewError(status int, code string) *Error {
	return &Error{status: status, code: code}
}

func (e *Error) Error() string { return e.code }

func (e *Error) StatusCode() int { return e.status }

func (e *Error) ErrorCode() string { return e.code }

var (
	ErrUnauthorized     = NewError(401, "unauthorized")
	ErrMissing          = NewError(404, "missing")
	ErrInvalidJSON      = NewError(400, "invalid_json")
	ErrInvalidName      = NewError(400, "invalid_name")
	ErrInvalidKind      = NewError(400, "invalid_kind")
	ErrInvalidMetal     = NewError(400, "invalid_metal")
	ErrInvalidPurity    = NewError(400, "invalid_purity")
	ErrInvalidWeight    = NewError(400, "invalid_weight")
	ErrInvalidCost      = NewError(400, "invalid_cost")
	ErrInvalidCurrency  = NewError(400, "invalid_currency")
	ErrInvalidNotes     = NewError(400, "invalid_notes")
	ErrInvalidDate      = NewError(400, "invalid_date")
	ErrInvalidPrice     = NewError(400, "invalid_price")
	ErrInvalidImage     = NewError(400, "invalid_image")
	ErrImageUnavailable = NewError(503, "image_unavailable")
	ErrFutureSale       = NewError(400, "future_sale")
	ErrAlreadySold      = NewError(409, "already_sold")
	ErrStillInStock     = NewError(409, "still_in_stock")
	ErrSoldLocked       = NewError(409, "sold_locked")
)
