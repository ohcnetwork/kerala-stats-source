package main

import (
	"strings"
	"sync"
	"time"

	"log"

	. "scrape/common"
	"scrape/dhs"
	"scrape/scraper"
)

const (
	HISTORIES_FILE         = "./histories.json"
	LATEST_FILE            = "./latest.json"
	SUMMARY_FILE           = "./summary.json"
	TEST_REPORTS_FILE      = "./testreports.json"
	HOTSPOT_HISTORIES_FILE = "./hotspots_histories.json"
	HOTSPOT_LATEST_FILE    = "./hotspots.json"
)

type Histories struct {
	History     []scraper.History `json:"histories"`
	LastUpdated string            `json:"last_updated"`
}

type HotspotsHistories struct {
	History     []dhs.HotspotsHistory `json:"histories"`
	LastUpdated string                `json:"last_updated"`
}

type LatestHistory struct {
	Summary     map[string]scraper.DistrictInfo `json:"summary"`
	Delta       map[string]scraper.DistrictInfo `json:"delta"`
	LastUpdated string                          `json:"last_updated"`
}

type LatestHotspotsHistory struct {
	Hotspots    []dhs.Hotspots `json:"hotspots"`
	LastUpdated string         `json:"last_updated"`
}

type Summary struct {
	Summary     scraper.DistrictInfo `json:"summary"`
	Delta       scraper.DistrictInfo `json:"delta"`
	LastUpdated string               `json:"last_updated"`
}

type TestReports struct {
	Reports     []scraper.TestReport `json:"reports"`
	LastUpdated string               `json:"last_updated"`
}

var (
	date        string
	lastUpdated string
	wg          sync.WaitGroup
)

func handleHistories() {
	defer wg.Done()
	var histories Histories
	ReadJSON(HISTORIES_FILE, &histories)
	last := len(histories.History) - 1
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
	WriteJSON(histories, HISTORIES_FILE)
	log.Println("histories written")
	latestData := LatestHistory{Summary: b.Summary, Delta: b.Delta, LastUpdated: lastUpdated}
	WriteJSON(latestData, LATEST_FILE)
	s, d := scraper.LatestSummary(b)
	log.Println("latest written")
	summary := Summary{Summary: s, Delta: d, LastUpdated: lastUpdated}
	WriteJSON(summary, SUMMARY_FILE)
	log.Println("summary written")
}

func handleTestReports() {
	defer wg.Done()
	var testReports TestReports
	ReadJSON(TEST_REPORTS_FILE, &testReports)
	last := len(testReports.Reports) - 1
	if date == testReports.Reports[last].Date {
		testReports.Reports[last] = scraper.ScrapeTodaysTestReport(date)
		log.Println("test report replaced")
	} else {
		testReports.Reports = append(testReports.Reports, scraper.ScrapeTodaysTestReport(date))
		log.Println("test report appended")
	}
	WriteJSON(testReports, TEST_REPORTS_FILE)
	log.Println("test reports written")
}

func handleHotspotsHistories() {
	defer wg.Done()
	var hhistories HotspotsHistories
	ReadJSON(HOTSPOT_HISTORIES_FILE, &hhistories)
	last := len(hhistories.History) - 1
	var hh dhs.HotspotsHistory
	if date == hhistories.History[last].Date {
		hh = dhs.ParseHotspotHistory(date)
		hhistories.History[last] = hh
		log.Println("hotspot history replaced")
	} else {
		hh = dhs.ParseHotspotHistory(date)
		hhistories.History = append(hhistories.History, hh)
		log.Println("hotspot history appended")
	}
	hhistories.LastUpdated = lastUpdated
	WriteJSON(hhistories, HOTSPOT_HISTORIES_FILE)
	log.Println("hotspots histories written")
	latestHotspotData := LatestHotspotsHistory{Hotspots: hh.Hotspots, LastUpdated: lastUpdated}
	WriteJSON(latestHotspotData, HOTSPOT_LATEST_FILE)
	log.Println("hotspots latest written")
}

func main() {
	log.Println("started")
	start := time.Now()
	lastUpdated = scraper.ScrapeLastUpdated()
	date = strings.Split(lastUpdated, " ")[0]
	wg.Add(1)
	go handleHistories()
	wg.Add(1)
	go handleTestReports()
	wg.Add(1)
	go handleHotspotsHistories()
	wg.Wait()
	log.Printf("completed in %v", time.Now().Sub(start))
}
