package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	rss "github.com/mattn/go-pkg-rss"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/gopher"
)

var (
	file = flag.String("f", "", "list of rss/atom feeds")
)

func findGopher() *gopher.Gopher {
	gophers := gopher.Lookup()
	if len(gophers) == 0 {
		return nil
	}
	return gophers[rand.Int()%len(gophers)]
}

func notify(item *rss.Item) {
	for _, link := range item.Links {
		var err error
		if gopher := findGopher(); gopher != nil {
			err = gopher.Message(item.Title, link.Href)
		}
		if err != nil {
			title := runewidth.Truncate(item.Title, 79, "...")
			fmt.Fprintf(color.Output, "%s\n\t%s\n",
				color.YellowString("%v", title),
				color.MagentaString("%v", link.Href))
		}
	}
}

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	uris := []string{}

	uris = append(uris, flag.Args()...)
	if *file != "" {
		b, err := ioutil.ReadFile(*file)
		if err != nil {
			fmt.Fprintln(os.Stderr, os.Args[0]+":", err)
			os.Exit(1)
		}
		for _, line := range strings.Split(string(b), "\n") {
			uris = append(uris, strings.TrimSpace(line))
		}
	}

	if len(uris) == 0 {
		uris = []string{"http://feeds.feedburner.com/hatena/b/hotentry"}
	}
	for {
		for _, uri := range uris {
			err := rss.New(5, true,
				func(feed *rss.Feed, newchannels []*rss.Channel) {
					fmt.Fprintf(color.Output, "%d new channel(s) in %s\n",
						len(newchannels), color.GreenString("%v", feed.Url))
				},
				func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
					fmt.Fprintf(color.Output, "%d new item(s) in %s\n",
						len(newitems), color.GreenString("%v", feed.Url))
					for _, item := range newitems {
						notify(item)
					}
				},
			).Fetch(uri, nil)

			if err != nil {
				fmt.Fprintf(os.Stderr, "[e] %s: %s", uri, err)
			}
			time.Sleep(10 * time.Second)
		}
		time.Sleep(3 * time.Minute)
	}
}
