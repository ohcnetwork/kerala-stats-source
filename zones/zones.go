package zones

import (
	"encoding/json"
	"io/ioutil"
	"log"
	. "scrape/common"
	"strings"
)

const ERROR_MSG = "error getting latest district zone information"

type Districts map[string]string

type Zones struct {
	Districts Districts `json:"districts"`
	Date      string    `json:"date"`
}

func GetDistictZones(date string) Zones {
	zones := Zones{Districts: make(map[string]string), Date: date}
	res, code := MakeRequest("https://api.covid19india.org/zones.json")
	if code != 200 {
		log.Panicln(ERROR_MSG)
	}
	defer res.Close()
	data, err := ioutil.ReadAll(res)
	if err != nil {
		log.Panicln(ERROR_MSG)
	}
	var z struct {
		Zones []struct {
			District     string `json:"district"`
			Districtcode string `json:"districtcode"`
			Lastupdated  string `json:"lastupdated"`
			Source       string `json:"source"`
			State        string `json:"state"`
			Statecode    string `json:"statecode"`
			Zone         string `json:"zone"`
		} `json:"zones"`
	}
	err = json.Unmarshal(data, &z)
	if err != nil {
		log.Panicln(ERROR_MSG)
	}
	for _, z := range z.Zones {
		if z.State == "Kerala" {
			zones.Districts[z.District] = strings.ToLower(z.Zone)
		}
	}
	if len(zones.Districts) != 14 {
		log.Panicln(ERROR_MSG)
	}
	return zones
}
