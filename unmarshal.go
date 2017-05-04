package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type segment struct {
	image string
	date  string
	time  string
}

func (s segment) String() string {
	return fmt.Sprintf("%s (%s | %s)", s.image, s.date, s.time)
}

type segments []segment

func (s segments) String() (out string) {
	for _, v := range s {
		out += fmt.Sprintf("%s\n", v.String())
	}
	return
}

func getSegments(url string) segments {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	reg := regexp.MustCompile("(\\/dynamic.*\\.png).*(\\d{2}\\.\\d{2}\\.\\d{4}).*(\\d{2}\\:\\d{2})")
	matches := reg.FindAllStringSubmatch(string(data), 100)
	segs := make(segments, len(matches))
	for i, v := range matches {
		segs[i].image = v[1]
		segs[i].date = v[2]
		segs[i].time = v[3]
	}
	return segs
}
