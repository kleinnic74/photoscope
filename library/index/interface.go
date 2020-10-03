package index

import "bitbucket.org/kleinnic74/photos/library"

type Name string

type Status int

const (
	NotIndexed = Status(iota)
	Indexed
	ErrorOnIndex
)

type State map[Name]Status

func NewState() State {
	return make(State)
}

func (s State) StatusFor(index Name) Status {
	status, _ := s[index]
	return status
}

func (s State) Set(index Name, status Status) {
	s[index] = status
}

type Tracker interface {
	RegisterIndex(Name)
	Update(Name, library.PhotoID, error) error
	Get(library.PhotoID) (State, bool, error)
	GetMissingIndexes(library.PhotoID) ([]Name, error)
}
