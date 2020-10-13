package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
)

func Atoi(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
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
	j = bytes.ReplaceAll(j, []byte("\\n"), []byte(""))
	re := regexp2.MustCompile(`\s{2,}`, 0)
	a, err := re.Replace(string(j), " ", -1, -1)
	if err != nil {
		log.Panicln(err)
	}
	j = []byte(a)
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, j); err != nil {
		log.Panicln(err)
	}
	err = ioutil.WriteFile(filename, buffer.Bytes(), 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func MakeRequest(url string) (io.ReadCloser, int, error) {
	res, err := http.Get(url)
	if err != nil {
		return &io.PipeReader{}, 0, err
	}
	return res.Body, res.StatusCode, nil
}
