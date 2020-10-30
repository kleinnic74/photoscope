package index

import (
	"encoding/json"

	"bitbucket.org/kleinnic74/photos/library"
)

type Name string

func (n Name) String() string {
	return string(n)
}

type Status int

const (
	NotIndexed = Status(iota)
	Indexed
	ErrorOnIndex
)

type IndexStatus struct {
	Status  Status          `json:"status"`
	Version library.Version `json:"version"`
}

func (s *IndexStatus) UnmarshalJSON(data []byte) error {
	// For backward compatibility, used to be a single int
	var newStruct struct {
		Status  Status          `json:"status"`
		Version library.Version `json:"version"`
	}
	if err := json.Unmarshal(data, &newStruct); err == nil {
		s.Status = newStruct.Status
		s.Version = newStruct.Version
		return nil
	}
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	s.Status = status
	s.Version = 1
	return nil
}

type State map[Name]IndexStatus

func NewState() State {
	return make(State)
}

func (s State) StatusFor(index Name) IndexStatus {
	status, _ := s[index]
	return status
}

func (s State) Set(index Name, status Status, version library.Version) {
	s[index] = IndexStatus{status, version}
}

type Tracker interface {
	RegisterIndex(Name, library.Version)
	Update(Name, library.PhotoID, error) error
	Get(library.PhotoID) (State, bool, error)
	GetMissingIndexes(library.PhotoID) ([]Name, error)
}
