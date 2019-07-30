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
)

var t_site = "https://xray.fm"
var t_shows = t_site + "/shows"

// fetch url, returning either a parsed goquery Document or an error.
func fetch(url string) (*goquery.Document, error) {
	var doc *goquery.Document

	res, err := http.Get(url)
	if err != nil {
		return doc, errors.Wrapf(err, "Failed to fetch: `%s'", url)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return doc, errors.Errorf("`%s' error: %s", url, res.Status)
	}

	// Load the document
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return doc, errors.Wrapf(err, "Failed to parse `%s'", url)
	}

	return doc, nil
}

// Given a url to an xray show, create a Feed from it
func make_feed(url string) (feeds.Feed, error) {
	var feed feeds.Feed

	doc, err := fetch(url)
	if err != nil {
		return feed, errors.Wrap(err, "Failed to download show page")
	}

	feed.Title = doc.Find("h1.main-title").Text()
	feed.Author = &feeds.Author{Name: doc.Find("div.hosts-container").Text()}
	feed.Link = &feeds.Link{Href: url}

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
		feed.Items = append(feed.Items, &item)
	})

	return feed, nil
}

func make_item(url string) (feeds.Item, error) {
	var itm feeds.Item

	doc, err := fetch(url)
	if err != nil {
		return itm, errors.Wrapf(err, "Failed to get episode page `%s'", url)
	}

	node := doc.Find("a.player")
	enclosure_url, exists := node.Attr("href")
	if !exists {
		return itm, errors.New("Couldn't find a.player href attr")
	}

	itm.Title = node.Text()
	itm.Link = &feeds.Link{Href: url, Rel: "canonical"}
	itm.Enclosure = &feeds.Enclosure{
		Url:    enclosure_url,
		Length: "unknown",
		Type:   "audio/mpeg",
	}
	itm.Content = doc.Find("div.creek-playlist").Text()

	return itm, nil
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
