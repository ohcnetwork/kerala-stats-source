package scraper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/eregnier/sffuzzy"
)

var sessid = ""

var districtMap = map[string]string{
	"TVM": "Thiruvananthapuram",
	"KLM": "Kollam",
	"PTA": "Pathanamthitta",
	"ALP": "Alappuzha",
	"KTM": "Kottayam",
	"IDK": "Idukki",
	"EKM": "Ernakulam",
	"TSR": "Thrissur",
	"PKD": "Palakkad",
	"MPM": "Malappuram",
	"KKD": "Kozhikode",
	"WYD": "Wayanad",
	"KNR": "Kannur",
	"KGD": "Kasaragod"}

var districtList = []string{"Thiruvananthapuram", "Kollam", "Pathanamthitta", "Alappuzha", "Kottayam", "Idukki", "Ernakulam", "Thrissur", "Palakkad", "Malappuram", "Kozhikode", "Wayanad", "Kannur", "Kasaragod"}

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

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Panicln(err)
	}
	return i
}

func getDoc(source string, referer string) goquery.Document {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", source, nil)
	req.Host = "dashboard.kerala.gov.in"
	req.Header.Set("Origin", "https://dashboard.kerala.gov.in")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:75.0) Gecko/20100101 Firefox/75.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", referer)
	req.Header.Set("Connection", "keep-alive")
	if sessid != "" {
		req.Header.Set("Cookie", sessid)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Panicf("status code error: %d %s", res.StatusCode, res.Status)
	}
	if sessid == "" {
		sessid = strings.Split(res.Header.Get("Set-Cookie"), ";")[0]
	}
	// if source == "https://dashboard.kerala.gov.in/quarantined-datewise.php" {
	// 	s, _ := ioutil.ReadAll(res.Body)
	// 	ioutil.WriteFile("test", s, 644)
	// }
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Panicln(err)
	}
	return *doc
}

func getDoc2(source string) goquery.Document {
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
		log.Panicln(err)
	}
	client := &http.Client{}
	var req *http.Request
	req, err = http.NewRequest("POST", "https://phantomjscloud.com/api/browser/v2/ak-kgg4y-v3ekt-rrzj0-5ab60-6wxh0/", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	// if res.StatusCode != 200 {
	// 	log.Panicf("status code error: %d %s", res.StatusCode, res.Status)
	// }
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Panicln(err)
	}
	return *doc
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
		Total:    atoi(row[1]),
		Negative: atoi(row[2]),
		Positive: atoi(row[3]),
		Pending:  atoi(row[4]),
		Today:    atoi(row[5]),
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
			data[districtMap[row[0]]] = row[1:]
			row = nil
		})
	})
	return data
}

func scrapeTable2(doc goquery.Document, selector string) map[string][]string {
	var row []string
	data := make(map[string][]string)
	doc.Find(selector).Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			data[districtMap[row[0]]] = row[1:]
			row = nil
		})
	})
	return data
}

func ocr() map[string][]string {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", "https://dashboard.kerala.gov.in/index.php", nil)
	req.Host = "dashboard.kerala.gov.in"
	req.Header.Set("Origin", "https://dashboard.kerala.gov.in")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:75.0) Gecko/20100101 Firefox/75.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", "https://dashboard.kerala.gov.in/index.php")
	req.Header.Set("Connection", "keep-alive")
	if sessid != "" {
		req.Header.Set("Cookie", sessid)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	s, _ := ioutil.ReadAll(res.Body)
	re := regexp.MustCompile(`geojson/.*center.geojson`)
	li := re.FindString(string(s))
	sessid = strings.Split(res.Header.Get("Set-Cookie"), ";")[0]
	client = &http.Client{}
	req, _ = http.NewRequest("GET", "https://dashboard.kerala.gov.in/"+li, nil)
	req.Host = "dashboard.kerala.gov.in"
	req.Header.Set("Origin", "https://dashboard.kerala.gov.in")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:75.0) Gecko/20100101 Firefox/75.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Set("Referer", "https://dashboard.kerala.gov.in/index.php")
	req.Header.Set("Connection", "keep-alive")
	if sessid != "" {
		req.Header.Set("Cookie", sessid)
	}
	res, err = client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	s, _ = ioutil.ReadAll(res.Body)

	type Foo struct {
		Crs struct {
			Properties struct {
				Name string `json:"name"`
			} `json:"properties"`
			Type string `json:"type"`
		} `json:"crs"`
		Features []struct {
			Geometry struct {
				Coordinates []float64 `json:"coordinates"`
				Type        string    `json:"type"`
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
	var foo Foo
	err = json.Unmarshal(s, &foo)
	if err != nil {
		log.Panic(err)
	}
	data := make(map[string][]string)
	for _, v := range foo.Features {
		p := v.Properties
		data[sffuzzy.SearchOnce(p.District, &districtList, sffuzzy.Options{Sort: true, Limit: 2, Normalize: true}).Results[0].Target] = []string{itoa(p.CovidStat), itoa(p.CovidStatcured), itoa(p.CovidStatdeath), itoa(p.CovidStatactive)}
	}
	return data
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ScrapeTodaysHistory(today string, last History) History {
	start := time.Now()
	// url1 := "https://dashboard.kerala.gov.in/dailyreporting.php"
	url2 := "https://dashboard.kerala.gov.in/quarantined-datewise.php"

	data1 := ocr()
	// data1 := scrapeTable2(getDoc2(url1), ".table > tbody:nth-child(2)")
	if len(data1) < 1 {
		log.Panicln("error scraping table1")
	}
	data2 := scrapeTable(getDoc(url2, url2), "table.table:nth-child(1) > tbody:nth-child(3)")
	if len(data2) < 1 {
		log.Panicln("error scraping table2")
	}
	b := History{Summary: make(map[string]DistrictInfo), Delta: make(map[string]DistrictInfo), Date: today}
	for _, d := range districtMap {
		b.Summary[d] = DistrictInfo{
			Confirmed:           atoi(data1[d][0]),
			Recovered:           atoi(data1[d][1]),
			Deceased:            atoi(data1[d][2]),
			Active:              atoi(data1[d][3]),
			TotalObservation:    atoi(data2[d][0]),
			HospitalObservation: atoi(data2[d][1]),
			HomeObservation:     atoi(data2[d][2]),
			HospitalizedToday:   atoi(data2[d][3]),
		}
		b.Delta[d] = DistrictInfo{
			Confirmed:           atoi(data1[d][0]) - last.Summary[d].Confirmed,
			Recovered:           atoi(data1[d][1]) - last.Summary[d].Recovered,
			Deceased:            atoi(data1[d][2]) - last.Summary[d].Deceased,
			Active:              atoi(data1[d][3]) - last.Summary[d].Active,
			TotalObservation:    atoi(data2[d][0]) - last.Summary[d].TotalObservation,
			HospitalObservation: atoi(data2[d][1]) - last.Summary[d].HospitalObservation,
			HomeObservation:     atoi(data2[d][2]) - last.Summary[d].HomeObservation,
			HospitalizedToday:   atoi(data2[d][3]) - last.Summary[d].HospitalizedToday,
		}
	}
	log.Printf("scraped today's history (%v) in %v\n", today, time.Now().Sub(start))
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
