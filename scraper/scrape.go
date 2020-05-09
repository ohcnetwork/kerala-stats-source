package scraper

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"log"

	. "scrape/common"

	"github.com/PuerkitoBio/goquery"
)

type DistrictInfo struct {
	HospitalObservation int `json:"hospital_obs"`
	HomeObservation     int `json:"home_obs"`
	TotalObservation    int `json:"total_obs"`
	HospitalizedToday   int `json:"hospital_today"`
	Confirmed           int `json:"confirmed"`
	Recovered           int `json:"recovered"`
	Deceased            int `json:"deceased"`
	Active              int `json:"active"`
}

type History struct {
	Summary map[string]DistrictInfo `json:"summary"`
	Delta   map[string]DistrictInfo `json:"delta"`
	Date    string                  `json:"date"`
}

type TestReport struct {
	Date     string `json:"date"`
	Total    int    `json:"total"`
	Negative int    `json:"negative"`
	Positive int    `json:"positive"`
	Pending  int    `json:"pending"`
	Today    int    `json:"today"`
}

func ScrapeLastUpdated() string {
	s := ""
	url := "https://dashboard.kerala.gov.in/index.php"
	doc := getDoc(url, url)
	s = doc.Find(".breadcrumb-item").Text()
	s = strings.ToUpper(strings.TrimSpace(strings.Split(s, ": ")[1]))
	if s == "" {
		log.Panicln("error scraping last updated")
	}
	return s
}

func ScrapeTodaysTestReport(today string) TestReport {
	start := time.Now()
	doc := getDoc(
		"https://dashboard.kerala.gov.in/testing-view-public.php",
		"https://dashboard.kerala.gov.in/quar_dst_wise_public.php",
	)
	var found *goquery.Selection
	var b *TestReport
	var row []string
	re := regexp.MustCompile(`\d\d-\d\d-\d\d\d\d`)
	firstrow := doc.Find(".table > tbody:nth-child(3)").Children()
	firstrow.EachWithBreak(func(indexth int, rowhtml *goquery.Selection) bool {
		if re.FindString(rowhtml.Text()) == today {
			found = rowhtml
			return false
		}
		return true
	})
	if found == nil {
		log.Panicln("no test report matching the date found")
	}
	found.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
		row = append(row, tablecell.Text())
	})
	b = &TestReport{
		Date:     row[0],
		Total:    Atoi(row[1]),
		Negative: Atoi(row[2]),
		Positive: Atoi(row[3]),
		Pending:  Atoi(row[4]),
		Today:    Atoi(row[5]),
	}
	log.Printf("scraped test reports in %v", time.Now().Sub(start))
	return *b
}

func scrapeTable(doc goquery.Document, selector string) map[string][]string {
	var row []string
	data := make(map[string][]string)
	doc.Find(selector).Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			data[DistrictMap[row[0]]] = row[1:]
			row = nil
		})
	})
	return data
}

func scrapeGeoJSON() map[string][]string {
	var body io.ReadCloser
	baseurl := "https://dashboard.kerala.gov.in/"
	body = makeRequest(baseurl+"index.php", baseurl+"index.php")
	defer body.Close()
	s, err := ioutil.ReadAll(body)
	if err != nil {
		log.Panicln(err)
	}
	li := regexp.MustCompile(`maps/.*outside.geojson`).FindString(string(s))
	body = makeRequest(baseurl+li, baseurl+"index.php")
	defer body.Close()
	s, err = ioutil.ReadAll(body)
	if err != nil {
		log.Panicln(err)
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
		log.Panic(err)
	}
	data := make(map[string][]string)
	for _, v := range geoJSON.Features {
		p := v.Properties
		data[FuzzySearch(p.District, DistrictList).Match] = []string{Itoa(p.CovidStat), Itoa(p.CovidStatcured), Itoa(p.CovidStatdeath), Itoa(p.CovidStatactive)}
	}
	return data
}

func ScrapeTodaysHistory(today string, last History) History {
	start := time.Now()
	// url1 := "https://dashboard.kerala.gov.in/dailyreporting.php"
	url2 := "https://dashboard.kerala.gov.in/quarantined-datewise.php"

	data1 := scrapeGeoJSON()
	// data1 := scrapeTable(getDoc2(url1), ".table > tbody:nth-child(2)")
	if len(data1) < 1 {
		log.Panicln("error scraping table1")
	}
	data2 := scrapeTable(getDoc(url2, url2), "table.table:nth-child(1) > tbody:nth-child(3)")
	if len(data2) < 1 {
		log.Panicln("error scraping table2")
	}
	b := History{Summary: make(map[string]DistrictInfo), Delta: make(map[string]DistrictInfo), Date: today}
	for _, d := range DistrictMap {
		b.Summary[d] = DistrictInfo{
			Confirmed:           Atoi(data1[d][0]),
			Recovered:           Atoi(data1[d][1]),
			Deceased:            Atoi(data1[d][2]),
			Active:              Atoi(data1[d][3]),
			TotalObservation:    Atoi(data2[d][0]),
			HospitalObservation: Atoi(data2[d][1]),
			HomeObservation:     Atoi(data2[d][2]),
			HospitalizedToday:   Atoi(data2[d][3]),
		}
		b.Delta[d] = DistrictInfo{
			Confirmed:           Atoi(data1[d][0]) - last.Summary[d].Confirmed,
			Recovered:           Atoi(data1[d][1]) - last.Summary[d].Recovered,
			Deceased:            Atoi(data1[d][2]) - last.Summary[d].Deceased,
			Active:              Atoi(data1[d][3]) - last.Summary[d].Active,
			TotalObservation:    Atoi(data2[d][0]) - last.Summary[d].TotalObservation,
			HospitalObservation: Atoi(data2[d][1]) - last.Summary[d].HospitalObservation,
			HomeObservation:     Atoi(data2[d][2]) - last.Summary[d].HomeObservation,
			HospitalizedToday:   Atoi(data2[d][3]) - last.Summary[d].HospitalizedToday,
		}
	}
	log.Printf("scraped latest history (%v) in %v\n", today, time.Now().Sub(start))
	return b
}

func LatestSummary(h History) (DistrictInfo, DistrictInfo) {
	var pos, dis, act, det, tot, hos, home, tod, dpos, ddis, dact, ddet, dtot, dhos, dhome, dtod = 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
	for _, info := range h.Summary {
		pos += info.Confirmed
		dis += info.Recovered
		act += info.Active
		det += info.Deceased
		tot += info.TotalObservation
		hos += info.HospitalObservation
		home += info.HomeObservation
		tod += info.HospitalizedToday
	}
	for _, info := range h.Delta {
		dpos += info.Confirmed
		ddis += info.Recovered
		dact += info.Active
		ddet += info.Deceased
		dtot += info.TotalObservation
		dhos += info.HospitalObservation
		dhome += info.HomeObservation
		dtod += info.HospitalizedToday
	}
	summary := DistrictInfo{
		Confirmed:           pos,
		Recovered:           dis,
		Active:              act,
		Deceased:            det,
		HospitalObservation: hos,
		HomeObservation:     home,
		TotalObservation:    tot,
		HospitalizedToday:   tod,
	}
	delta := DistrictInfo{
		Confirmed:           dpos,
		Recovered:           ddis,
		Active:              dact,
		Deceased:            ddet,
		HospitalObservation: dhos,
		HomeObservation:     dhome,
		TotalObservation:    dtot,
		HospitalizedToday:   dtod,
	}
	return summary, delta
}
