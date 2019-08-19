// xpod creates podcast feeds for XRay.fm shows.
package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"time"
)

var t_site = "https://xray.fm"
var t_shows = t_site + "/shows"

// fetch url, returning either a parsed goquery Document or an error.
func fetch(url string) (*http.Response, *goquery.Document, error) {
	var doc *goquery.Document

	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to fetch: `%s'", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.Errorf("`%s' error: %s", url, resp.Status)
	}

	// Load the document
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to parse `%s'", url)
	}

	return resp, doc, nil
}

// Given a url to an xray show, create a Feed from it
func make_feed(url string) (*feeds.Feed, error) {
	_, doc, err := fetch(url)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to download show page")
	}

	feed := feeds.Feed{
		Title:  doc.Find("h1.main-title").Text(),
		Author: &feeds.Author{Name: doc.Find("div.hosts-container").Text()},
		Link:   &feeds.Link{Href: url},
	}

	// Find episodes
	doc.Find("div.broadcast.cfm-has-audio > div.info > div.title > a").Each(func(i int, s *goquery.Selection) {
		// Find the link and title
		link, exists := s.Attr("href")
		if !exists {
			log.Printf("Can't find episode href for `%s'", url)
			return // Skip
		}
		item, err := make_item(t_site + link)
		if err != nil {
			log.Printf("Couldn't make Item for `%s': %s", url, err)
			return // Skip
		}
		feed.Items = append(feed.Items, item)
	})

	return &feed, nil
}

// Parse a timestamp string into a time.Time
// If parsing fails, return the zero-value timestamp.
func parse_time(time_s string) (time.Time, error) {
	var ts time.Time

	ts, err := time.Parse("3:00pm, 1-2-2006", time_s)
	if err != nil {
		return ts, errors.Errorf("Failed to parse timestamp `%s'", time_s)
	}

	return ts, nil
}

// Create a feed Item from the URL to an episode of a show.
func make_item(url string) (*feeds.Item, error) {
	resp, doc, err := fetch(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get episode page `%s'", url)
	}

	node := doc.Find("a.player")
	enclosure_link, exists := node.Attr("href")
	if !exists {
		return nil, errors.New("Couldn't find a.player href attr")
	}

	enclosure_url := enclosure_link
	parsed_url, err := resp.Request.URL.Parse(enclosure_link)
	if err != nil {
		enclosure_url = parsed_url.String()
	}

	created, err := parse_time(doc.Find("div.date").Text())
	if err != nil {
		log.Printf("Falling back to nil timestamp: %s", err)
	}

	itm := feeds.Item{
		Title:   node.Text(),
		Created: created,
		Link:    &feeds.Link{Href: url, Rel: "canonical"},
		Enclosure: &feeds.Enclosure{
			Url:    enclosure_url,
			Length: "unknown",
			Type:   "audio/mpeg",
		},
		Content: doc.Find("div.creek-playlist").Text(),
	}

	return &itm, nil
}

func main() {
	exit_status := 0
	verbose := true

	for _, show := range os.Args[1:] {
		url := t_shows + "/" + show
		feed_file := show + ".rss"
		if verbose {
			fmt.Printf("%s -> %s\n", url, feed_file)
		}

		f, err := os.Create(feed_file)
		if err != nil {
			log.Printf("Failed to open `%s': %s", feed_file, err)
			exit_status++
			continue
		}

		feed, err := make_feed(url)
		if err != nil {
			log.Printf("Failed to generate feed: %s", err)
			exit_status++
			continue
		}

		feed.WriteRss(f) // Or .WriteAtom
		f.Close()
	}
	os.Exit(exit_status)
}
