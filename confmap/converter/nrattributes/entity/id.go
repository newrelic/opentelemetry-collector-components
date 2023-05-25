package entity

// ID entity ID
type ID int64

// GUID entity GUID
type GUID string

// Identity entity identifiers
type Identity struct {
	ID   ID
	GUID GUID
}

const (
	EmptyID   = ID(0)
	EmptyGUID = GUID("")
)

// IsEmpty returns if ID is empty
func (i ID) IsEmpty() bool {
	return i == EmptyID
}
