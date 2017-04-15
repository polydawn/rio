package rio

import "fmt"

// Grouping type for any errors raised by Rio.
// You probably want to use one of the more specific groupings to
// describe which parts of the system are the source of issues, though.
//
// All `rio.Error` are reasonably serializable as json.
type Error interface {
	rioError() // marker method.
	error
}

/*
	Raised when a `context.Context` cancelled a long operation part-way through.
*/
type Cancelled struct{}

func (Cancelled) rioError() {}
func (e Cancelled) Error() string {
	return "Cancelled"
}

/*
	Raised when encountering clearly corrupt contents read from a warehouse.

	This is distinct from `ErrHashMismatch` in that it represents some
	form of failure to parse data before we have even reached the stage
	where the content's full semantic hash is computable (for example,
	with a tar transmat, if the tar header is completely nonsense, we
	just have to give up).
*/
type ErrWareCorrupt struct {
	Msg    string        `json:"msg"`
	WareID WareID        `json:"wareID"`
	From   WarehouseAddr `json:"from"`
}

func (ErrWareCorrupt) rioError() {}
func (e ErrWareCorrupt) Error() string {
	return fmt.Sprintf("Ware Corrupt: %s, while working on %q from %s", e.Msg, e.WareID, e.From)
}
