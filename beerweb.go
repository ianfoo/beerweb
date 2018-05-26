// Experiment: Scrape the Chuck's Hop Shop beer list.
package beerweb

import (
	"strings"
)

// Taplister is an interface to be called and have returned a list of beers,
// allowing for multiple strategies like HTML table scraping and API access.
type Taplister interface {
	FetchBeers() ([]Beer, error)
	Venue() string
	URL() string // TODO This is slightly awkward.
}

// Beer describes a beer by brewery, name, and any other available attributes.
type Beer struct {
	Brewery string `json:"brewery"`
	Name    string `json:"name"`
	Style   string `json:"style"`
	ABV     string `json:"abv"`
	Origin  string `json:"origin"`
}

// Format the Beer as a pretty-ish string.
func (b Beer) String() string {
	s := strings.Builder{}
	s.WriteString(b.Brewery)
	if b.Origin != "" {
		s.WriteByte('(')
		s.WriteString(b.Origin)
		s.WriteByte(')')
	}
	s.WriteString(" | ")
	s.WriteString(b.Name)
	if b.Style != "" {
		s.WriteString(" | ")
		s.WriteString(b.Style)
	}
	if b.ABV != "" {
		s.WriteString(" | ")
		s.WriteString(b.ABV)
		if !strings.HasSuffix(b.ABV, "%") {
			s.WriteByte('%')
		}
		s.WriteString(" abv")
	}
	return s.String()
}

// Valid will return true if a beer has a brewery and a name.
// Extra details are gravy.
func (b Beer) Valid() bool {
	return b.Brewery != "" && b.Name != ""
}
