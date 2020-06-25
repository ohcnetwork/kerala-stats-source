package scraper

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"regexp"
	. "scrape/common"
)

func scrapeGeoJSON() (map[string][]string, error) {
	var body io.ReadCloser
	data := make(map[string][]string)
	baseurl := "https://dashboard.kerala.gov.in/"
	body, err := makeRequest(baseurl+"index.php", baseurl+"index.php")
	defer body.Close()
	if err != nil {
		return data, err
	}
	s, err := ioutil.ReadAll(body)
	if err != nil {
		return data, err
	}
	li := regexp.MustCompile(`maps/.*outside.geojson`).FindString(string(s))
	body, err = makeRequest(baseurl+li, baseurl+"index.php")
	defer body.Close()
	s, err = ioutil.ReadAll(body)
	if err != nil {
		return data, err
	}
	var geoJSON struct {
		Crs struct {
			Properties struct {
				Name string `json:"name"`
			} `json:"properties"`
			Type string `json:"type"`
		} `json:"crs"`
		Features []struct {
			Geometry struct {
				Coordinates [][][]float64 `json:"coordinates"`
				Type        string        `json:"type"`
			} `json:"geometry"`
			Properties struct {
				District        string `json:"District"`
				Objectid        int64  `json:"OBJECTID"`
				CovidStat       int64  `json:"covid_stat"`
				CovidStatactive int64  `json:"covid_statactive"`
				CovidStatcured  int64  `json:"covid_statcured"`
				CovidStatdeath  int64  `json:"covid_statdeath"`
			} `json:"properties"`
			Type string `json:"type"`
		} `json:"features"`
		Type string `json:"type"`
	}
	err = json.Unmarshal(s, &geoJSON)
	if err != nil {
		return data, err
	}
	for _, v := range geoJSON.Features {
		p := v.Properties
		data[FuzzySearch(p.District, DistrictList).Match] = []string{Itoa(p.CovidStat), Itoa(p.CovidStatcured), Itoa(p.CovidStatdeath), Itoa(p.CovidStatactive)}
	}
	return data, nil
}
