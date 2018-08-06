// A slapdash HTTP/HTML implementation to show the beer lists from the defined
// venues.  Certain aspects of this implementation are a little offensive. This
// should be corrected.
package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ianfoo/beerweb"
	"github.com/ianfoo/beerweb/venues"
)

var (
	addr = flag.String("addr", ":5050", "Address to listen on")
)

// TODO Do not use unassociated global variables to track the tap lists.
var (
	taplists []beerweb.Taplist
	mu       sync.RWMutex
)

func init() {
	if !strings.Contains(*addr, ":") {
		*addr = ":" + *addr
	}
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/", beerHandler)
	s := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		MaxHeaderBytes:    1 << 20,
		Handler:           m,
	}

	// Shut down the beer fetch operations.
	shutdownCh := make(chan struct{})
	s.RegisterOnShutdown(func() {
		shutdownCh <- struct{}{}
	})
	go getBeers(shutdownCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	shutdownFinished := make(chan struct{})
	go func() {
		log.Println("listening on", s.Addr)
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println("server error:", err)
			os.Exit(1)
		}
		log.Println("exiting")
		shutdownFinished <- struct{}{}
	}()

	sig := <-sigCh
	log.Println("shutting down after receiving signal", sig)
	err := s.Shutdown(context.Background())
	if err != nil {
		log.Println("error shutting down:", err)
	}
	<-shutdownFinished
}

const pollInterval = 10 * time.Minute

func getBeers(shutdown <-chan struct{}) {
	fetch := func() {
		var (
			t          = time.Now()
			totalBeers int
			err        error
		)
		defer func() {
			log.Printf(
				"fetched %d beers from %d venues in %v",
				totalBeers, len(taplists), time.Since(t))
		}()

		log.Println("fetching beers")
		newTaplists, err := beerweb.FetchAll(venues.Venues)
		if err != nil {
			if faErr, ok := err.(beerweb.FetchAllError); ok {
				log.Println("error fetching taplists:", faErr)
				return
			}
			panic(err)
		}

		mu.Lock()
		for _, tl := range newTaplists {
			// FIXME This is really ugly.
			// But I wanted a quick way to show that the beer lists had changed.
			// Proper taplist diffing will help this go away.
			var oldTaplist *beerweb.Taplist
			for _, otl := range taplists {
				if tl.Venue == otl.Venue {
					oldTaplist = &otl
					break
				}
			}
			if oldTaplist != nil && !reflect.DeepEqual(tl, oldTaplist) {
				log.Printf("beers for %s have changed!", tl.Venue)
			}
			totalBeers += len(tl.Beers)
		}
		taplists = newTaplists
		mu.Unlock()
	}
	fetch()

	t := time.NewTicker(pollInterval)
	for {
		select {
		case <-t.C:
			fetch()
		case <-shutdown:
			log.Println("exiting beer fetch goroutine")
			t.Stop()
			return
		}
	}
}

func addPercent(s string) string {
	if !pctPat.MatchString(s) {
		return s
	}
	if !strings.HasSuffix(s, "%") {
		return s + "%"
	}
	return s
}

var (
	pctPat = regexp.MustCompile("^[0-9.]+")
	tmpl   = template.Must(template.New("Taplists").
		Funcs(template.FuncMap{"percent": addPercent}).
		Parse(tmplStr))
)

func beerHandler(rw http.ResponseWriter, r *http.Request) {
	t := time.Now()
	defer log.Printf("serviced request from %s in %v", r.RemoteAddr, time.Since(t))
	mu.RLock()
	defer mu.RUnlock()
	tmpl.ExecuteTemplate(rw, "Taplists", taplists)
}

var semanticUICDN = `<link rel="stylesheet" type="text/css"` +
	`href="https://cdnjs.cloudflare.com/ajax/libs/semantic-ui/2.3.3/semantic.min.css"/>`

var tmplStr = `<!DOCTYPE html>
<html>` + semanticUICDN + `<head>
<title>Beer Lists</title>
</head>
<body>
<div class="ui header segment">
<div class="ui center aligned container">
<h1>Beer Lists</h1>
</div>
</div>
{{range $taplist := .}}
<div class="ui one column container">
<div class="column">

<table class="ui celled striped inverted compact table">
  <thead>
  <tr>
  <th colspan="5" class="ui">Beers at {{ $taplist.Venue }}</th>
  <tr>
    <th>Brewery</th>
    <th>Name</th>
    <th>Style</th>
    <th>ABV</th>
    <th>Origin</th>
  </tr>
  </thead>
  <tbody>
{{range $beer := $taplist.Beers}}
  <tr>
    <td>{{ $beer.Brewery }}</td>
    <td>{{ $beer.Name }}</td>
    <td>{{ $beer.Style }}</td>
    <td>{{ percent $beer.ABV }}</td>
    <td>{{ $beer.Origin }}</td>
  </tr>
{{end}}
</tbody>
</table>
</div>
</div>
<div class="ui hidden divider"></div>
{{end}}
<div class="ui footer segment">
<div class="ui center aligned container">
<p>Generated using <a href="https://github.com/ianfoo/beerweb">beerweb</a>.</p>
</div>
</div>
</body>
</html>`
