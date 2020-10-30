package main

import (
	"fmt"
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
	data, err := getter(url + "?nid=557")
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
