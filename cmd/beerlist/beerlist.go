package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/ianfoo/beerweb"
)

var sites = []beerweb.Taplist{
	{
		Venue: "Chuck's Hop Shop (Greenwood)",
		URL:   "http://chucks.jjshanks.net/draft",
		Processor: beerweb.MakeHTMLTableScraper(
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
		Processor: beerweb.MakeHTMLTableScraper(
			"table.taplist-table > tbody",
			"td:nth-child(2)",
			"td:nth-child(3)",
			"td:nth-child(4)",
			"td:nth-child(7)",
			"td:nth-child(8)"),
	},
}

func main() {
	jsonOutput := flag.Bool("json", false, "write output as JSON")
	flag.Parse()

	for i, tl := range sites {
		beers := []beerweb.Beer{}
		c := colly.NewCollector()
		tl.Processor(c, &beers)
		c.Visit(tl.URL)

		if *jsonOutput {
			output := map[string]interface{}{
				"Venue": tl.Venue,
				"URL":   tl.URL,
				"Beers": beers,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.Encode(output)
			continue
		}

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
