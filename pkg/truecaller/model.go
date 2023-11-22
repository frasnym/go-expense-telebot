package truecaller

import "fmt"

type SearchResponse struct {
	Name string `json:"name"`
	Raw  string
}

func (e *SearchResponse) ParseInformationMessage() string {
	header := "Information:"
	name := fmt.Sprintf("\nName: %s", e.Name)

	return fmt.Sprint(header, name)
}
