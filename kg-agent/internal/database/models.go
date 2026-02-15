package database

import "fmt"

type DocumentEntityResponse struct {
	Id    string
	Title string
}

func (d *DocumentEntityResponse) Print() string {
	return fmt.Sprintf("Document_id: %s - Title: %s", d.Id, d.Title)
}
