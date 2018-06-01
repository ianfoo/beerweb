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

func main() {
	jsonOutput := flag.Bool("json", false, "write output as JSON")
	flag.Parse()

	taplists, err := beerweb.FetchAll(venues.Venues)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(taplists)
		return
	}

	for i, taplist := range taplists {
		banner := "Beer list for " + taplist.Venue
		underline := strings.Repeat("=", len(banner))
		fmt.Printf("%s\n%s\n", banner, underline)
		for _, beer := range taplist.Beers {
			fmt.Println(beer)
		}
		if i < len(taplists)-1 {
			fmt.Println()
		}
	}

}
