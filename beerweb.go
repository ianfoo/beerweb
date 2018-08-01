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
	var s strings.Builder
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

// TextTable renders a list of beers entry in compliance with the configured
// field widths.
type TextTable struct {
	beers []Beer

	BreweryWidth,
	NameWidth,
	StyleWidth,
	ABVWidth,
	OriginWidth int
}

// Add saves a beer entry, and compares a its field widths to the
// currently-stored format widths, and updates any of them that are shorter
// that the current beer's field widths.
func (tt *TextTable) Add(b Beer) {
	tt.beers = append(tt.beers, b)
	if l := len([]rune(b.Brewery)); l > tt.BreweryWidth {
		tt.BreweryWidth = l
	}
	if l := len([]rune(b.Name)); l > tt.NameWidth {
		tt.NameWidth = l
	}
	if l := len([]rune(b.Style)); l > tt.StyleWidth {
		tt.StyleWidth = l
	}
	if l := len([]rune(b.ABV)); l > tt.ABVWidth {
		tt.ABVWidth = l
	}
	if l := len([]rune(b.Origin)); l > tt.OriginWidth {
		tt.OriginWidth = l
	}
}

// NewTextTable builds a model for a nicely-formatted text table based, given
// a slice of Beer values.
func NewTextTable(beers []Beer) *TextTable {
	table := &TextTable{}
	for _, beer := range beers {
		table.Add(beer)
	}
	return table
}

func (tt *TextTable) header() string {
	var s strings.Builder
	columns := []struct {
		name  string
		width *int
	}{
		{"Brewery", &tt.BreweryWidth},
		{"Name", &tt.NameWidth},
		{"Style", &tt.StyleWidth},
		{"ABV", &tt.ABVWidth},
		{"Origin", &tt.OriginWidth},
	}
	for _, column := range columns {
		if *column.width == 0 {
			continue
		}
		// Handle the case where the column name is longer than any
		// value present in the column. This is pretty ugly since we're
		// mutating the TextTable state inside of the header() func.
		// There's definitely a much better way to do this, but, later.
		if l := len(column.name); l > *column.width {
			*column.width = l
		}
		s.WriteString("| ")
		s.WriteString(column.name)
		s.WriteString(strings.Repeat(" ", *column.width-len([]rune(column.name))))
		s.WriteByte(' ')
	}
	s.WriteByte('|')
	return s.String()
}

// Format renders the fields of a Beer using a minimum width for each field,
// bordered by spaces and a pipe character. This is useful for writing a list
// of beers as a table that can be easily read.
func (tt TextTable) String() string {
	var s strings.Builder
	header := tt.header()
	underline := strings.Repeat("=", tt.Width())
	s.WriteString(underline)
	s.WriteByte('\n')
	s.WriteString(header)
	s.WriteByte('\n')
	s.WriteString(underline)
	s.WriteByte('\n')

	for _, b := range tt.beers {
		fields := []struct {
			value string
			width int
		}{
			{b.Brewery, tt.BreweryWidth},
			{b.Name, tt.NameWidth},
			{b.Style, tt.StyleWidth},
			{b.ABV, tt.ABVWidth},
			{b.Origin, tt.OriginWidth},
		}
		for _, field := range fields {
			if field.width == 0 {
				continue
			}
			s.WriteString("| ")
			s.WriteString(field.value)
			s.WriteString(strings.Repeat(" ", field.width-len([]rune(field.value))+1))
		}
		s.WriteString("|\n")
	}
	s.WriteString(underline)
	return s.String()
}

// Width returns the entire with of a formatted row, including space padding
// and separator characters.
func (tt TextTable) Width() int {
	numZeroWidth := 0
	if tt.BreweryWidth == 0 {
		numZeroWidth++
	}
	if tt.NameWidth == 0 {
		numZeroWidth++
	}
	if tt.StyleWidth == 0 {
		numZeroWidth++
	}
	if tt.ABVWidth == 0 {
		numZeroWidth++
	}
	if tt.OriginWidth == 0 {
		numZeroWidth++
	}

	width := tt.BreweryWidth +
		tt.NameWidth +
		tt.StyleWidth +
		tt.ABVWidth +
		tt.OriginWidth

	// account for padding, pipe separator chars, and fields that won't be rendered
	width += 16 - numZeroWidth*3
	return width
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
	// We need to explicitly return nil for the error value here if there have
	// been no issues, since the error variable declared at the top of this
	// function will fail a comparison to nil, even if there were no errors
	// encountered during the fetch.
	if len(err) == 0 {
		return taplists, nil
	}
	return taplists, err
}

type FetchAllError []error

func (e FetchAllError) Error() string {
	var b strings.Builder
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
