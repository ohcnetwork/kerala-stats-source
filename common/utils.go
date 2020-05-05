package common

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
)

func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Panicln(err)
	}
	return i
}

func Itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func FuzzySearch(search string, targets []string) *fuzzy.MatchPair {
	d, err := fuzzy.ExtractOne(search, targets)
	if err != nil {
		log.Panicln(err)
	}
	return d
}

func ReadJSON(filename string, v interface{}) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal([]byte(file), v)
	if err != nil {
		log.Panic(err)
	}
}

func WriteJSON(v interface{}, filename string) {
	j, err := json.Marshal(v)
	if err != nil {
		log.Panicln(err)
	}
	err = ioutil.WriteFile(filename, j, 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func MakeRequest(url string) (io.ReadCloser, int) {
	res, err := http.Get(url)
	if err != nil {
		log.Panicln(err)
	}
	return res.Body, res.StatusCode
}
