package scraper

import (
	"github.com/PuerkitoBio/goquery"
)

func getDoc(source string, referer string) (*goquery.Document, error) {
	var doc *goquery.Document
	body, err := makeRequest(source, referer)
	defer body.Close()
	if err != nil {
		return doc, err
	}
	doc, err = goquery.NewDocumentFromReader(body)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
