package cursor

type Page struct {
	Data interface{} `json:"data"`
	Prev string      `json:"prev,omitempty"`
	Next string      `json:"next,omitempty"`
}

func PageFor(data interface{}, cursor Cursor) (page Page) {
	page.Data = data
	if previous, exists := cursor.Previous(); exists {
		page.Prev = previous.Encode()
	}
	if next, exists := cursor.Next(); exists {
		page.Next = next.Encode()
	}
	return
}

func Unpaged(data interface{}) (page Page) {
	return Page{Data: data}
}
