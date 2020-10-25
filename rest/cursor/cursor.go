package cursor

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"bitbucket.org/kleinnic74/photos/consts"
)

type Cursor struct {
	Start    int
	PageSize int
	Order    consts.SortOrder `json:"order,omitempty"`
}

var defaultPageSize int = 20

func DecodeFromRequest(r *http.Request) Cursor {
	cursor := Cursor{PageSize: defaultPageSize}
	encodedCursor := r.Form.Get("c")
	if err := DecodeFromString(encodedCursor, &cursor); err != nil {
		// log a warning
	}
	switch pageSizeStr := r.Form.Get("p"); pageSizeStr {
	case "":
	default:
		if pageSize64, err := strconv.ParseUint(pageSizeStr, 10, 0); err == nil {
			cursor.PageSize = int(pageSize64)
		}
	}
	switch order := r.Form.Get("o"); order {
	case "a", "+", "asc":
		cursor.Order = consts.Ascending
	case "d", "-", "desc":
		cursor.Order = consts.Descending
	default:
	}
	return cursor
}

func DecodeFromString(encoded string, cursor *Cursor) error {
	if encoded == "" {
		return nil
	}
	asJSON, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(asJSON, cursor)
}

func (c Cursor) Encode() string {
	asJSON, err := json.Marshal(&c)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(asJSON))
}

func (c Cursor) Previous() (Cursor, bool) {
	if c.Start > 0 {
		return Cursor{Start: c.Start - c.PageSize, PageSize: c.PageSize, Order: c.Order}, true
	}
	return Cursor{}, false
}

func (c Cursor) Next() (Cursor, bool) {
	return Cursor{Start: c.Start + c.PageSize, PageSize: c.PageSize, Order: c.Order}, true
}
