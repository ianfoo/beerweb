// package beerweb defines the types necessary for fetching and displaying
// beer lists.
package beerweb

import (
	"fmt"
	"strings"
)

// Taplister is an interface to be called and have returned a list of beers,
// allowing for multiple strategies like HTML table scraping and API access.
type Taplister interface {
	FetchBeers() ([]Beer, error)
	Venue() string
	URL() string // TODO This is slightly awkward.
}

type Taplist struct {
	Venue string `json:"venue"`
	URL   string `json:"url"`
	Beers []Beer `json:"beers"`
}

// Beer describes a beer by brewery, name, and any other available attributes.
type Beer struct {
	Brewery string `json:"brewery"`
	Name    string `json:"name"`
	Style   string `json:"style"`
	ABV     string `json:"abv"`
	Origin  string `json:"origin"`
}

// String formats the Beer as a pretty-ish string.
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

// FetchAll is a helper to concurrently get all the beers for
// a slice of Taplisters.
//
// TODO This is kind of ugly, but it allows us to use a common function
// for fetching all venue's beers for all clients (e.g., CLI, web).
func FetchAll(venues []Taplister) ([]Taplist, error) {
	respCh := make(chan response)
	taplists := make([]Taplist, 0, len(venues))

	var err FetchAllError

	defer func() { close(respCh) }()

	for _, tl := range venues {
		go fetch(tl, respCh)
	}

	for i := len(venues); i > 0; i-- {
		tl := <-respCh
		if tl.err != nil {
			err = append(err, tl.err)
			continue
		}
		taplists = append(taplists, Taplist{
			Venue: tl.venue,
			URL:   tl.url,
			Beers: tl.beers,
		})
	}
	return taplists, err
}

type FetchAllError []error

func (e FetchAllError) Error() string {
	b := strings.Builder{}
	for _, err := range e {
		b.WriteString(err.Error())
		b.WriteByte('\n')
	}
	return b.String()
}

// TODO This can be simplified
type response struct {
	venue string
	url   string
	beers []Beer
	err   error
}

func fetch(tl Taplister, ch chan<- response) {
	beers, err := tl.FetchBeers()
	if err != nil {
		ch <- response{err: fmt.Errorf("error fetching beers from %s: %v\n", tl.Venue(), err)}
		return
	}
	ch <- response{
		venue: tl.Venue(),
		url:   tl.URL(),
		beers: beers,
	}
}
