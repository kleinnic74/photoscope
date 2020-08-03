package cursor

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
)

type Cursor struct {
	Start    uint
	PageSize uint
}

var defaultPageSize uint = 20

func DecodeFromRequest(r *http.Request) Cursor {
	cursor := Cursor{PageSize: defaultPageSize}
	pageSizeStr := r.URL.Query().Get("p")
	var pageSize uint
	if pageSizeStr != "" {
		if pageSize64, err := strconv.ParseUint(pageSizeStr, 10, 0); err == nil {
			pageSize = uint(pageSize64)
			cursor.PageSize = pageSize
		}
	}
	encodedCursor := r.URL.Query().Get("c")
	cursor = DecodeFromString(encodedCursor, defaultPageSize)
	if pageSize != 0 {
		cursor.PageSize = pageSize
	}
	return cursor
}

func DecodeFromString(encoded string, defaultPageSize uint) Cursor {
	cursor := Cursor{PageSize: defaultPageSize}
	if encoded == "" {
		return cursor
	}
	asJSON, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return cursor
	}
	if err := json.Unmarshal(asJSON, &cursor); err != nil {
		return cursor
	}
	if cursor.PageSize == 0 {
		cursor.PageSize = defaultPageSize
	}
	return cursor
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
		return Cursor{Start: c.Start - c.PageSize, PageSize: c.PageSize}, true
	}
	return Cursor{}, false
}

func (c Cursor) Next() (Cursor, bool) {
	return Cursor{Start: c.Start + c.PageSize, PageSize: c.PageSize}, true
}
