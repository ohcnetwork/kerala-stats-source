package main

import (
	"strings"
	"sync"
	"time"

	"log"

	. "scrape/common"
	"scrape/scraper"
	"scrape/zones"
)

const (
	HISTORIES_FILE         = "./histories.json"
	LATEST_FILE            = "./latest.json"
	SUMMARY_FILE           = "./summary.json"
	TEST_REPORTS_FILE      = "./testreports.json"
	HOTSPOT_HISTORIES_FILE = "./hotspots_histories.json"
	HOTSPOT_LATEST_FILE    = "./hotspots.json"
	ZONES_HISTORIES_FILE   = "./zones_histories.json"
	ZONES_LATEST_FILE      = "./zones.json"
)

type Histories struct {
	History     []scraper.History `json:"histories"`
	LastUpdated string            `json:"last_updated"`
}

type HotspotsHistories struct {
	History     []scraper.HotspotsHistory `json:"histories"`
	LastUpdated string                    `json:"last_updated"`
}

type ZoneHistories struct {
	History     []zones.Zones `json:"histories"`
	LastUpdated string        `json:"last_updated"`
}

type LatestHistory struct {
	Summary     map[string]scraper.DistrictInfo `json:"summary"`
	Delta       map[string]scraper.DistrictInfo `json:"delta"`
	LastUpdated string                          `json:"last_updated"`
}

type LatestHotspotsHistory struct {
	Hotspots    []scraper.Hotspots `json:"hotspots"`
	LastUpdated string             `json:"last_updated"`
}

type LatestZones struct {
	Districts   zones.Districts `json:"districts"`
	LastUpdated string          `json:"last_updated"`
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
	var err error
	var histories Histories
	ReadJSON(HISTORIES_FILE, &histories)
	last := len(histories.History) - 1
	var b scraper.History
	if date == histories.History[last].Date {
		b, err = scraper.ScrapeTodaysHistory(date, histories.History[last-1])
		if err != nil {
			log.Println("ERROR scraping todays history", err)
			return
		}
		histories.History[last] = b
		log.Println("history replaced")
	} else {
		b, err = scraper.ScrapeTodaysHistory(date, histories.History[last])
		if err != nil {
			log.Println("ERROR scraping todays history", err)
			return
		}
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
	latest, err := scraper.ScrapeTodaysTestReport(date)
	if err != nil {
		log.Println("ERROR scraping todays test reports", err)
		return
	}
	if date == testReports.Reports[last].Date {
		testReports.Reports[last] = latest
		log.Println("test report replaced")
	} else {
		testReports.Reports = append(testReports.Reports, latest)
		log.Println("test report appended")
	}
	testReports.LastUpdated = lastUpdated
	WriteJSON(testReports, TEST_REPORTS_FILE)
	log.Println("test reports written")
}

func handleHotspotsHistories() {
	defer wg.Done()
	var hhistories HotspotsHistories
	ReadJSON(HOTSPOT_HISTORIES_FILE, &hhistories)
	last := len(hhistories.History) - 1
	hh, err := scraper.ScrapeHotspotsHistory(date)
	if err != nil {
		log.Println("ERROR getting hotspots histories", err)
		return
	}
	if date == hhistories.History[last].Date {
		hhistories.History[last] = hh
		log.Println("hotspot history replaced")
	} else {
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

func handleZonesHistories() {
	defer wg.Done()
	var zhistories ZoneHistories
	ReadJSON(ZONES_HISTORIES_FILE, &zhistories)
	last := len(zhistories.History) - 1
	zz, err := zones.GetDistictZones(date)
	if err != nil {
		log.Println("ERROR getting zones histories", err)
		return
	}
	if date == zhistories.History[last].Date {
		zhistories.History[last] = zz
		log.Println("zones history replaced")
	} else {
		zhistories.History = append(zhistories.History, zz)
		log.Println("zones history appended")
	}
	zhistories.LastUpdated = lastUpdated
	WriteJSON(zhistories, ZONES_HISTORIES_FILE)
	log.Println("zones histories written")
	latestZones := LatestZones{Districts: zz.Districts, LastUpdated: lastUpdated}
	WriteJSON(latestZones, ZONES_LATEST_FILE)
	log.Println("zones latest written")
}

func main() {
	var err error
	log.Println("started")
	start := time.Now()
	lastUpdated, err = scraper.ScrapeLastUpdated()
	if err != nil {
		log.Panicln("ERROR getting last updated", err)
	}
	log.Printf("last updated on %v", lastUpdated)
	date = strings.Split(lastUpdated, " ")[0]
	wg.Add(1)
	go handleHistories()
	wg.Add(1)
	go handleHotspotsHistories()
	wg.Add(1)
	go handleZonesHistories()
	wg.Add(1)
	go handleTestReports()
	wg.Wait()
	log.Printf("completed in %v", time.Now().Sub(start))
}
