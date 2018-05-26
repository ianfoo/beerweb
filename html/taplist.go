package html

import (
	"github.com/gocolly/colly"
	"github.com/ianfoo/beerweb"
)

type Taplist struct {
	venue string
	url   string
	Beers []beerweb.Beer

	collector *colly.Collector
}

type TaplistConfig struct {
	Venue           string
	URL             string
	TableSelector   string
	BrewerySelector string
	NameSelector    string
	StyleSelector   string
	OriginSelector  string
	ABVSelector     string
}

// TODO Actually return errors
func (tl *Taplist) FetchBeers() ([]beerweb.Beer, error) {
	tl.collector.Visit(tl.url)
	return tl.Beers, nil
}

func (tl Taplist) Venue() string {
	return tl.venue
}

func (tl Taplist) URL() string {
	return tl.url
}

// makeHTMLTableScraper assumes that beers are listed in an HTML table,
// an returns a function that extract them, given HTML selectors to find
// the beer list and the beer details within each row.
func NewTaplist(coll *colly.Collector, c TaplistConfig) *Taplist {
	tl := &Taplist{
		collector: coll.Clone(),
		venue:     c.Venue,
		url:       c.URL,
	}

	tl.collector.OnHTML(c.TableSelector, func(table *colly.HTMLElement) {
		table.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			beer := beerweb.Beer{
				Brewery: row.ChildText(c.BrewerySelector),
				Name:    row.ChildText(c.NameSelector),
				Style:   row.ChildText(c.StyleSelector),
				Origin:  row.ChildText(c.OriginSelector),
				ABV:     row.ChildText(c.ABVSelector),
			}
			if beer.Valid() {
				tl.Beers = append(tl.Beers, beer)
			}
		})
	})

	return tl
}
