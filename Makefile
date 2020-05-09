SHELL := bash
loc=$(echo ./checkpoints/$(date '+%d-%m-%Y')/)

build:
	GOOS="linux" GOARCH="amd64" go build -o ../kerala_stats/scrape

data:
	cp -f ../kerala_stats/*.json .
	