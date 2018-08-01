package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ianfoo/beerweb"
	"github.com/ianfoo/beerweb/venues"
)

func main() {
	jsonOutput := flag.Bool("json", false, "write output as JSON")
	flag.Parse()

	taplists, err := beerweb.FetchAll(venues.Venues)
	if err != nil {
		log.SetFlags(0)
		log.Fatalln("error fetching beer lists:", err)
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(taplists)
		return
	}

	for i, taplist := range taplists {
		fmt.Println("Beer list for " + taplist.Venue)
		fmt.Println(beerweb.NewTextTable(taplist.Beers))
		if i < len(taplists)-1 {
			fmt.Println()
		}
	}
}
