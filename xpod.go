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

func fetch(url string) (*goquery.Document, error) {
	var doc *goquery.Document

	res, err := http.Get(url)
	if err != nil {
		return doc, errors.Wrapf(err, "Failed to fetch: `%s'", url)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return doc, errors.Errorf("`%s' status code error: %d %s", url, res.StatusCode, res.Status)
	}

	// Load the document
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return doc, errors.Wrapf(err, "Failed to parse `%s'", url)
	}
	return doc, nil
}

func make_feed(url string) (feeds.Feed, error) {
	var feed feeds.Feed

	doc, err := fetch(url)
	if err != nil {
		return feed, errors.Wrapf(err, "Failed to fetch `%s'", url)
	}

	feed.Title = doc.Find("h1.main-title").Text()
	feed.Author = &feeds.Author{Name: doc.Find("div.hosts-container").Text()}
	feed.Link = &feeds.Link{Href: url}

	// Find shows
	doc.Find("div.broadcast.cfm-has-audio > div.info > div.title > a").Each(func(i int, s *goquery.Selection) {
		// Find the link and title
		link, _ := s.Attr("href")
		item, _ := make_item(t_site + link)
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
	enclosure_url, _ := node.Attr("href")

	itm.Title = node.Text()
	itm.Link = &feeds.Link{Href: url, Rel: "canonical"}
	itm.Enclosure = &feeds.Enclosure{Url: enclosure_url}
	itm.Content = doc.Find("div.creek-playlist").Text()

	return itm, nil
}

func main() {
	shows := []string{"heavy-metal-sewing-circle", "gothique-boutique", "sfutf"}

	for _, show := range shows {
		url := t_shows + "/" + show
		feed_file := show + ".xml"
		fmt.Printf("%s -> %s\n", url, feed_file)

		f, err := os.Create(feed_file)
		if err != nil {
			log.Printf("Failed to open `%s': %s", feed_file, err)
			continue
		}

		feed, err := make_feed(url)
		if err != nil {
			log.Printf("Failed to generate feed for `%s': %s", url, err)
			continue
		}

		feed_s, err := feed.ToAtom()
		if err != nil {
			log.Printf("Failed to generate feed for `%s': %s", url, err)
			continue
		}

		f.WriteString(feed_s)
		f.Close()
	}
}
