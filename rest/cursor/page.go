package cursor

type Link struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type Page struct {
	Data  interface{} `json:"data"`
	Links []Link      `json:"links,omitempty"`
}

func PageFor(data interface{}, cursor Cursor, hasMore bool) (page Page) {
	page.Data = data
	if previous, exists := cursor.Previous(); exists {
		page.Links = append(page.Links, Link{"previous", previous.Encode()})
	}
	if next, exists := cursor.Next(); exists && hasMore {
		page.Links = append(page.Links, Link{"next", next.Encode()})
	}
	return
}

func Unpaged(data interface{}) (page Page) {
	return Page{Data: data}
}
