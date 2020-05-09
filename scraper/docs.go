package scraper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

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

func getDoc2(source string) (*goquery.Document, error) {
	var doc *goquery.Document
	tmp := strings.Split(sessid, "=")
	cookie := map[string]string{"name": tmp[0], "value": tmp[1]}
	type Data struct {
		Url             string              `json:"url"`
		RenderType      string              `json:"renderType"`
		UrlSettings     map[string]string   `json:"urlSettings"`
		Cookies         []map[string]string `json:"cookies"`
		RequestSettings map[string]string   `json:"requestSettings"`
		OutputAsJson    bool                `json:"outputAsJson"`
	}
	data := Data{
		Url:        source,
		RenderType: "html",
		UrlSettings: map[string]string{
			"operation": "GET",
			"encoding":  "utf8",
		},
		Cookies: []map[string]string{cookie},
		RequestSettings: map[string]string{
			"selector": "#wrapper2 > table > tbody:nth-child(2)",
		},
		OutputAsJson: false,
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		return doc, err
	}
	client := &http.Client{}
	var req *http.Request
	req, err = http.NewRequest("POST", "https://phantomjscloud.com/api/browser/v2/ak-kgg4y-v3ekt-rrzj0-5ab60-6wxh0/", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return doc, err
	}
	defer res.Body.Close()
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
