package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"log"

	"github.com/coronasafe/kerala_stats/scraper"
)

type Histories struct {
	History     []scraper.History `json:"histories"`
	LastUpdated string            `json:"last_updated"`
}

type TestReports struct {
	Reports     []scraper.TestReport `json:"reports"`
	LastUpdated string               `json:"last_updated"`
}

type LatestHistory struct {
	Summary     map[string]scraper.DistrictInfo `json:"summary"`
	Delta       map[string]scraper.DistrictInfo `json:"delta"`
	LastUpdated string                          `json:"last_updated"`
}

type Summary struct {
	Summary     scraper.DistrictInfo `json:"summary"`
	Delta       scraper.DistrictInfo `json:"delta"`
	LastUpdated string               `json:"last_updated"`
}

func writeJSON(v interface{}, filename string) {
	j, err := json.Marshal(v)
	if err != nil {
		log.Panicln(err)
	}
	err = ioutil.WriteFile(filename, j, 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func main() {
	log.Println("started")
	start := time.Now()
	lastUpdated := scraper.ScrapeLastUpdated()
	var histories Histories
	file, err := ioutil.ReadFile("./histories.json")
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal([]byte(file), &histories)
	if err != nil {
		log.Panic(err)
	}
	last := len(histories.History) - 1
	date := strings.Split(lastUpdated, " ")[0]
	var b scraper.History
	if date == histories.History[last].Date {
		b = scraper.ScrapeTodaysHistory(date, histories.History[last-1])
		histories.History[last] = b
		log.Println("history replaced")
	} else {
		b = scraper.ScrapeTodaysHistory(date, histories.History[last])
		histories.History = append(histories.History, b)
		log.Println("history appended")
	}
	histories.LastUpdated = lastUpdated
	writeJSON(histories, "./histories.json")
	latestData := LatestHistory{Summary: b.Summary, Delta: b.Delta, LastUpdated: lastUpdated}
	writeJSON(latestData, "./latest.json")
	s, d := scraper.LatestSummary(b)
	log.Println("latests written")
	summary := Summary{Summary: s, Delta: d, LastUpdated: lastUpdated}
	writeJSON(summary, "./summary.json")
	log.Println("summary written")

	var testReports TestReports
	file, err = ioutil.ReadFile("./testreports.json")
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal([]byte(file), &testReports)
	if err != nil {
		log.Panic(err)
	}
	last = len(testReports.Reports) - 1
	if date == testReports.Reports[last].Date {
		testReports.Reports[last] = scraper.ScrapeTodaysTestReport(date)
		log.Println("test reports replaced")
	} else {
		testReports.Reports = append(testReports.Reports, scraper.ScrapeTodaysTestReport(date))
		log.Println("test reports appended")
	}
	writeJSON(testReports, "./testreports.json")
	log.Printf("completed in %v", time.Now().Sub(start))
}
