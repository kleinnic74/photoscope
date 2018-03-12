package formats

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	"bitbucket.org/kleinnic74/photos/domain/gps"
)

const (
	fType     = "ftyp"
	movieData = "moov"
	meta      = "meta"
)

var (
	NotAQuickTimeFile error = fmt.Errorf("Not a quicktime file")
	atoms             map[string]AtomParser
	metaDataParsers   = map[string]MetaDataParser{
		"com.apple.quicktime.creationdate":     setCreationDate,
		"com.apple.quicktime.location.ISO6709": setLocation,
	}
)

type AtomParser func(*Quicktime, io.Reader, AtomContainer) error

type MetaDataParser func(*Quicktime, string, string) error

type AtomContainer interface {
	Add(a *Atom)
}

type AtomWalker func(*Atom, int)

type AtomWalkable interface {
	Walk(AtomWalker)
}

type Atom struct {
	typ   string
	size  uint64
	hsize uint64

	nested []*Atom
}

func (a *Atom) SizeOfData() int64 {
	return int64(a.size - a.hsize)
}

func (a *Atom) TypeName() string {
	return a.typ
}

type Quicktime struct {
	keys         map[uint32]string
	Atoms        []*Atom
	creationDate time.Time
	coords       gps.Coordinates
}

func (parent *Atom) Walk(f AtomWalker, level int) {

	f(parent, level)
	for _, child := range parent.nested {
		child.Walk(f, level+1)
	}
}

func (parent *Atom) Add(a *Atom) {
	parent.nested = append(parent.nested, a)
}

func ReadAsQuicktime(in io.Reader) (*Quicktime, error) {
	var qt Quicktime = Quicktime{
		Atoms: make([]*Atom, 0),
		keys:  make(map[uint32]string),
	}
	err := parseAtoms(&qt, in, &qt)
	if err != nil {
		return nil, err
	}
	return &qt, nil
}

func (qt *Quicktime) Add(a *Atom) {
	qt.Atoms = append(qt.Atoms, a)
}

func (qt *Quicktime) Walk(f AtomWalker) {
	for _, a := range qt.Atoms {
		a.Walk(f, 0)
	}
}

func (qt *Quicktime) DateTaken() time.Time {
	return qt.creationDate
}

func (qt *Quicktime) Location() *gps.Coordinates {
	return &qt.coords
}

func (qt *Quicktime) defineKey(i uint32, name string) {
	qt.keys[i] = name
}

func (qt *Quicktime) setMetaDataAsString(key, value string) {
	metaParser, found := metaDataParsers[key]
	if found {
		err := metaParser(qt, key, value)
		if err != nil {
			log.Printf("Could not parse meta-data %s: %s", key, err)
		}
	} else {
		log.Printf("  Ignored meta-data %s=%s (no parser defined)", key, value)
	}
}

func parseAtoms(qt *Quicktime, in io.Reader, parent AtomContainer) error {
	for a, err := nextAtom(in); err != io.EOF; a, err = nextAtom(in) {
		if err != nil {
			return err
		}
		parent.Add(a)
		content := io.LimitReader(in, a.SizeOfData())
		if parser, found := atoms[a.typ]; found {
			if parser != nil {
				if err = parser(qt, content, a); err != nil {
					return err
				}
			}
		} else {
			if err := skip(content, a.SizeOfData()); err != nil {
				return err
			}
		}
	}
	return nil
}

func skip(in io.Reader, nb int64) error {
	var total int64 = 0
	buf := make([]byte, 4096)
	for count, err := in.Read(buf); total < nb; count, err = in.Read(buf) {
		total += int64(count)
		if total != nb && err != nil {
			return err
		}
	}
	return nil
}

func parseKeys(qt *Quicktime, in io.Reader, parent AtomContainer) error {
	var header struct {
		VersionFlags uint32
		NbEntries    uint32
	}
	if err := binary.Read(in, binary.BigEndian, &header); err != nil {
		return err
	}
	var i uint32
	for i = 1; i <= header.NbEntries; i++ {
		var keySpec struct {
			Size      uint32
			Namespace uint32
		}
		if err := binary.Read(in, binary.BigEndian, &keySpec); err != nil {
			return err
		}
		var value []byte = make([]byte, keySpec.Size-8)
		if _, err := in.Read(value); err != nil {
			return err
		}
		qt.defineKey(i, string(value))
	}
	return nil
}

func parseItemList(qt *Quicktime, in io.Reader, parent AtomContainer) error {
	for a, err := nextAtom(in); err != io.EOF; a, err = nextAtom(in) {
		content := io.LimitReader(in, a.SizeOfData())
		index := binary.BigEndian.Uint32([]byte(a.typ))
		key, found := qt.keys[index]
		if found {
			if err := parseKeyValue(qt, key, content); err != nil {
				return err
			}
		} else {
			skip(content, a.SizeOfData())
		}
	}
	return nil
}

func parseKeyValue(qt *Quicktime, key string, in io.Reader) error {
	for a, err := nextAtom(in); err != io.EOF; a, err = nextAtom(in) {
		switch a.TypeName() {
		case "data":
			var dataSpec struct {
				Typ    uint32
				Locale uint32
			}
			if err := binary.Read(in, binary.BigEndian, &dataSpec); err != nil {
				return err
			}
			payloadSize := a.SizeOfData() - 8
			switch dataSpec.Typ {
			case 1:
				var buf []byte = make([]byte, payloadSize)
				if _, err := in.Read(buf); err != nil {
					return err
				}
				qt.setMetaDataAsString(key, string(buf))
			default:
				skip(in, payloadSize)
			}
			return nil
		default:
			skip(in, a.SizeOfData())
		}
	}
	return nil
}

func nextAtom(in io.Reader) (*Atom, error) {
	var a struct {
		Size uint32
		Typ  [4]byte
	}
	if err := binary.Read(in, binary.BigEndian, &a); err != nil {
		return nil, err
	}
	if a.Size == 1 {
		var buf []byte = make([]byte, 8)
		_, err := in.Read(buf)
		if err != nil {
			return nil, err
		}
		extsize := binary.BigEndian.Uint64(buf)
		return &Atom{
			size:  extsize,
			hsize: 16,
			typ:   string(a.Typ[0:4]),
		}, nil
	} else {
		return &Atom{
			size:  uint64(a.Size),
			hsize: 8,
			typ:   string(a.Typ[0:4]),
		}, nil
	}
}

func setCreationDate(qt *Quicktime, key, value string) error {
	creationDate, err := time.Parse("2006-01-02T15:04:05-0700", value)
	if err != nil {
		return err
	}
	qt.creationDate = creationDate
	return nil
}

func setLocation(qt *Quicktime, key, value string) error {
	re := regexp.MustCompile("([-+]\\d+\\.?\\d*)")
	matches := re.FindAllString(value, 3)
	if matches != nil && len(matches) >= 2 {
		lat, err := strconv.ParseFloat(matches[0], 64)
		if err != nil {
			return nil
		}
		long, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return nil
		}
		qt.coords = gps.NewCoordinates(lat, long)
	}
	return nil
}

func init() {
	atoms = map[string]AtomParser{
		movieData: parseAtoms,
		meta:      parseAtoms,
		"keys":    parseKeys,
		"ilst":    parseItemList,
	}
}
