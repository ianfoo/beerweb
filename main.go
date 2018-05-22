// Experiment: Scrape the Chuck's Hop Shop (85th) beer list.
package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
)

type Taplist struct {
	Venue     string
	URL       string
	processor func(*colly.Collector, *[]Beer)
}

var sites = []Taplist{
	{
		Venue: "Chuck's Hop Shop (Greenwood)",
		URL:   "http://chucks.jjshanks.net/draft",
		processor: makeHTMLTableScraper(
			"div[id=draft_list] > table",
			"td.draft_brewery",
			"td.draft_name",
			"",
			"td.draft_origin",
			"td.draft_abv"),
	},
	{
		Venue: "Chuck's Hop Shop (Central District)",
		URL:   "http://chuckstaplist.com",
		processor: makeHTMLTableScraper(
			"table.taplist-table > tbody",
			"td:nth-child(2)",
			"td:nth-child(3)",
			"td:nth-child(4)",
			"td:nth-child(7)",
			"td:nth-child(8)"),
	},
}

type Beer struct {
	Brewery string
	Name    string
	Style   string
	ABV     string
	Origin  string
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

func main() {
	for i, tl := range sites {
		beers := []Beer{}
		c := colly.NewCollector()
		tl.processor(c, &beers)
		c.Visit(tl.URL)

		banner := "Beer list for " + tl.Venue
		underline := strings.Repeat("=", len(banner))
		fmt.Printf("%s\n%s\n", banner, underline)
		for _, beer := range beers {
			fmt.Println(beer)
		}
		if i < len(sites)-1 {
			fmt.Println()
		}
	}

}

// makeHTMLTableScraper assumes that beers are listed in an HTML table,
// an returns a function that extract them, given HTML selectors to find
// the beer list and the beer details within each row.
func makeHTMLTableScraper(
	tableSelector,
	brewerySelector,
	nameSelector,
	styleSelector,
	originSelector,
	abvSelector string) func(c *colly.Collector, beers *[]Beer) {

	return func(c *colly.Collector, beers *[]Beer) {
		c.OnHTML(tableSelector, func(table *colly.HTMLElement) {
			table.ForEach("tr", func(_ int, row *colly.HTMLElement) {
				beer := Beer{
					Brewery: row.ChildText(brewerySelector),
					Name:    row.ChildText(nameSelector),
					Style:   row.ChildText(styleSelector),
					Origin:  row.ChildText(originSelector),
					ABV:     row.ChildText(abvSelector),
				}
				if beer.Valid() {
					*beers = append(*beers, beer)
				}
			})
		})
	}
}
