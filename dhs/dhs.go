package dhs

import (
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
	"time"

	"scrape/common"
	. "scrape/common"

	"github.com/PuerkitoBio/goquery"
	"github.com/lu4p/cat"
)

const (
	FEATURE_FILE = "./data/features.json"
)

type Hotspots struct {
	District string `json:"district"`
	LSGD     string `json:"lsgd"`
}

type HotspotsHistory struct {
	Hotspots []Hotspots `json:"hotspots"`
	Date     string     `json:"date"`
}

var (
	re1 = regexp.MustCompile(`Sl. (\012){0,1}(No.* ){0,1}District .*(\012No){0,1}`)
	re2 = regexp.MustCompile(`\d{1,3}\s\s[a-zA-Z]{5,}\s\s[a-zA-Z]+.*\012`)
	re3 = regexp.MustCompile(`((Municipality)|(Panchayat)|(Corporation))`)
	re4 = regexp.MustCompile(`\s*\n`)
	re5 = regexp.MustCompile(`\d{2}-\d{2}-\d{4}`)
	re6 = regexp.MustCompile(`\s*(\(.\))\s*`)
)

func GetBulletinPost(date string) string {
	url := "https://dhs.kerala.gov.in/category/daily-bulletin/"
	var s []byte
	var err error
	var link string
	i := 1
	d := strings.Split(date, "-")
	re := regexp.MustCompile(`/` + path.Join(d[2], d[1], d[0], date) + `(-2)*/`)
	for {
		body, code := MakeRequest(url)
		if code == 404 {
			log.Panicln("error finding the bulletin post for the date")
		}
		defer body.Close()
		s, err = ioutil.ReadAll(body)
		if err != nil {
			log.Panicln(err)
		}
		link = re.FindString(string(s))
		if link != "" {
			break
		}
		i++
		url = "https://dhs.kerala.gov.in/category/daily-bulletin/" + "page/" + Itoa(int64(i)) + "/"
	}
	return "https://dhs.kerala.gov.in" + link
}

func GetPDFURL(url string) string {
	body, code := MakeRequest(url)
	defer body.Close()
	if code != 200 {
		log.Panicln("error retrieving bulletin post")
	}
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Panicln(err)
	}
	s, exists := doc.Find(".entry-content > p:nth-child(1) > a:nth-child(1)").Attr("href")
	if !exists {
		s, exists = doc.Find(".entry-content > ul:nth-child(1) > li:nth-child(1) > a:nth-child(1)").Attr("href")
		if !exists {
			log.Panicln("error finding the pdf in the bulletin post")
		}
	}
	return "https://dhs.kerala.gov.in" + s
}

func DownloadPDF(date string) []byte {
	url := GetBulletinPost(date)
	log.Printf("retrieving pdf from url: %v", url)
	pdfurl := GetPDFURL(url)
	body, code := MakeRequest(pdfurl)
	defer body.Close()
	if code != 200 {
		log.Panicln("error downloading the pdf")
	}
	s, err := ioutil.ReadAll(body)
	if err != nil {
		log.Panicln(err)
	}
	return s
}

func ParseHotspotHistory(today string) HotspotsHistory {
	start := time.Now()
	history := HotspotsHistory{Hotspots: make([]Hotspots, 0), Date: today}
	txt, err := cat.FromBytes(DownloadPDF(today))
	if err != nil {
		log.Panicln(err)
	}
	data := re2.FindAllString(re1.Split(txt, 2)[1], -1)
	for _, l := range data {
		place := strings.Split(re4.ReplaceAllString(re3.ReplaceAllString(l, ""), ""), "  ")
		if place[2] == "Koothuparamba (M)" {
			place[2] = "Kuthuparambu (M)"
		}
		if place[2] == "Changanacherry (M)" {
			place[2] = "Changanassery (M)"
		}
		if place[2] == "District Hospital" {
			place[2] = "Marutharoad"
		}
		d := FuzzySearch(place[1], common.DistrictList)
		s := FuzzySearch(place[2], GeoLSG[d.Match])
		if s.Score < 60 {
			log.Panicln(place[2] + s.Match)
		}
		history.Hotspots = append(history.Hotspots, Hotspots{District: d.Match, LSGD: s.Match})
	}
	log.Printf("parsed latest hotspot history (%v) in %v with %v entries\n", today, time.Now().Sub(start), len(history.Hotspots))
	return history
}
