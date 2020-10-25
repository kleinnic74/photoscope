package cursor

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/logging"
	"go.uber.org/zap"
)

type Cursor struct {
	Start    int
	PageSize int
	Order    consts.SortOrder `json:"order,omitempty"`
}

var (
	defaultPageSize int = 20
	encoding            = base64.StdEncoding.WithPadding(base64.NoPadding)
)

func DecodeFromRequest(r *http.Request) Cursor {
	cursor := Cursor{PageSize: defaultPageSize}
	encodedCursor := r.URL.Query().Get("c")
	if err := DecodeFromString(encodedCursor, &cursor); err != nil {
		logging.From(r.Context()).Warn("Invalid cursor", zap.String("cursor", encodedCursor), zap.Error(err))
	} else {
		logging.From(r.Context()).Info("Received cursor", zap.String("cursor", encodedCursor), zap.Int("start", cursor.Start), zap.Int("page", cursor.PageSize))
	}
	switch pageSizeStr := r.URL.Query().Get("p"); pageSizeStr {
	case "":
	default:
		if pageSize64, err := strconv.ParseUint(pageSizeStr, 10, 0); err == nil {
			cursor.PageSize = int(pageSize64)
		}
	}
	switch order := r.URL.Query().Get("o"); order {
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
	asJSON, err := encoding.DecodeString(encoded)
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
	return encoding.EncodeToString([]byte(asJSON))
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
