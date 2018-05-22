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

// Sites to visit. Just Chuck's to start.
// Adding more will require different scraper functions, or API access.
var sites = []Taplist{
	{
		Venue:     "Chuck's Hop Shop (Greenwood)",
		URL:       "http://chucks.jjshanks.net/draft",
		processor: readBeerChucksGreenwood,
	},
	{
		Venue:     "Chuck's Hop Shop (Central District)",
		URL:       "http://chuckstaplist.com",
		processor: readBeerChucksCentralDistrict,
	},
}

type Beer struct {
	Brewery string
	Name    string
	Style   string
	ABV     string
	Origin  string
}

func (b Beer) String() string {
	return fmt.Sprintf("%s (%s) | %s | %s | %s%% abv",
		b.Brewery,
		b.Origin,
		b.Name,
		b.Style,
		b.ABV)
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

func readBeerChucksGreenwood(c *colly.Collector, beers *[]Beer) {
	c.OnHTML(`div[id=draft_list] > table`, func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			if !strings.HasPrefix(row.Attr("class"), "draft_") {
				return
			}
			beer := Beer{
				Brewery: row.ChildText("td.draft_brewery"),
				Name:    row.ChildText("td.draft_name"),
				ABV:     row.ChildText("td.draft_abv"),
				Origin:  row.ChildText("td.draft_origin"),
			}
			*beers = append(*beers, beer)
		})
	})
}

func readBeerChucksCentralDistrict(c *colly.Collector, beers *[]Beer) {
	c.OnHTML("table.taplist-table > tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			beer := Beer{
				Brewery: row.ChildText("td:nth-child(2)"),
				Name:    row.ChildText("td:nth-child(3)"),
				Style:   row.ChildText("td:nth-child(4)"),
				Origin:  row.ChildText("td:nth-child(7)"),
				ABV:     strings.Replace(row.ChildText("td:nth-child(8)"), "%", "", -1),
			}
			*beers = append(*beers, beer)
		})
	})
}
