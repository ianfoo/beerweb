package venues

import (
	"github.com/gocolly/colly"
	"github.com/ianfoo/beerweb"
	"github.com/ianfoo/beerweb/html"
)

var coll = colly.NewCollector()

var Venues = []beerweb.Taplister{
	html.NewTaplist(coll, html.TaplistConfig{
		Venue:           "Chuck's Hop Shop (Greenwood)",
		URL:             "http://chucks.jjshanks.net/draft",
		TableSelector:   "div[id=draft_list] > table",
		BrewerySelector: "td.draft_brewery",
		NameSelector:    "td.draft_name",
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
