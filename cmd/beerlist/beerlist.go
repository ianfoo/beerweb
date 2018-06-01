package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ianfoo/beerweb"
	"github.com/ianfoo/beerweb/venues"
)

type response struct {
	venue    string
	url      string
	beerList []beerweb.Beer
	err      error
}

func main() {
	jsonOutput := flag.Bool("json", false, "write output as JSON")
	flag.Parse()

	respCh := make(chan response)
	defer func() { close(respCh) }()

	for _, tl := range venues.Venues {
		go fetch(tl, respCh)
	}

	for i := 0; i < len(venues.Venues); i++ {
		resp := <-respCh
		if resp.err != nil {
			fmt.Println(resp.err)
			continue
		}

		if *jsonOutput {
			output := map[string]interface{}{
				"Venue": resp.venue,
				"URL":   resp.url,
				"Beers": resp.beerList,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.Encode(output)
			continue
		}

		banner := "Beer list for " + resp.venue
		underline := strings.Repeat("=", len(banner))
		fmt.Printf("%s\n%s\n", banner, underline)
		for _, beer := range resp.beerList {
			fmt.Println(beer)
		}
		if i < len(venues.Venues)-1 {
			fmt.Println()
		}
	}

}

func fetch(tl beerweb.Taplister, ch chan<- response) {
	beers, err := tl.FetchBeers()
	if err != nil {
		ch <- response{err: fmt.Errorf("error fetching beers from %s: %v\n", tl.Venue, err)}
		return
	}
	ch <- response{
		venue:    tl.Venue(),
		url:      tl.URL(),
		beerList: beers,
	}
}
