package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/ianfoo/beerweb"
	"github.com/ianfoo/beerweb/html"
	"os"
	"strings"
)

var coll = colly.NewCollector()

var sites = []beerweb.Taplister{
	html.NewTaplist(coll, html.TaplistConfig{
		Venue:           "Chuck's Hop Shop (Greenwood)",
		URL:             "http://chucks.jjshanks.net/draft",
		TableSelector:   "div[id=draft_list] > table",
		BrewerySelector: "td.draft_brewery",
		NameSelector:    "td.draft_name",
		StyleSelector:   "",
		OriginSelector:  "td.draft_origin",
		ABVSelector:     "td.draft_abv",
	}),
	html.NewTaplist(coll, html.TaplistConfig{
		Venue:           "Chuck's Hop Shop (Central District)",
		URL:             "http://chuckstaplist.com",
		TableSelector:   "table.taplist-table > tbody",
		BrewerySelector: "td:nth-child(2)",
		NameSelector:    "td:nth-child(3)",
		StyleSelector:   "td:nth-child(4)",
		OriginSelector:  "td:nth-child(7)",
		ABVSelector:     "td:nth-child(8)",
	}),
}

func main() {
	jsonOutput := flag.Bool("json", false, "write output as JSON")
	flag.Parse()

	for i, tl := range sites {
		beers, err := tl.FetchBeers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching beers from %s: %v\n", tl.Venue, err)
			continue
		}

		if *jsonOutput {
			output := map[string]interface{}{
				"Venue": tl.Venue(),
				"URL":   tl.URL(),
				"Beers": beers,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.Encode(output)
			continue
		}

		banner := "Beer list for " + tl.Venue()
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
