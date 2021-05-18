package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var cookies []*http.Cookie

const (
	precip = "https://www.meteo.lv/laiks/nokrisni/"
	clouds = "https://www.meteo.lv/laiks/makoni/"
	temper = "https://www.meteo.lv/laiks/temperatura/"
	confor = "https://www.meteo.lv/laiks/komforta-temperatura/"
	windin = "https://www.meteo.lv/laiks/vejs/"
)

var modes = []string{
	precip,
	clouds,
	temper,
	confor,
	windin,
}

var current = 0

func main() {
	var err error
	cookies, err = getCookie("https://www.meteo.lv/")
	if err != nil {
		panic(err)
	}
	segment := getSegments(precip)
	b, err := walk.NewBitmapFromImageForDPI(getImage(segment[0]), 96)
	fatal(err)
	image := &walk.ImageView{}
	selector := &walk.ComboBox{}

	MainWindow{
		Title: "Meteo",
		Size: Size{
			Width:  650,
			Height: 400},
		Layout: VBox{},
		Children: []Widget{
			ComboBox{Model: []string{
				"Nokrišņi",
				"Mākoņi",
				"Temperatūra",
				"Konforta temperatūra",
				"Vējš",
			},
				CurrentIndex: 0,
				OnCurrentIndexChanged: func() {
					id := selector.CurrentIndex()
					segment = getSegments(modes[id])
					b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
					fatal(err)
					image.SetImage(b)
				},
				AssignTo: &selector,
			},
			ImageView{
				Image:    b,
				AssignTo: &image,
			},
			HSplitter{
				Children: []Widget{
					PushButton{
						Text: "<<<",
						OnClicked: func() {
							if current-2 < 0 {
								current = len(segment) - 1
							} else {
								current -= 3
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: "<<",
						OnClicked: func() {
							if current-2 < 0 {
								current = len(segment) - 1
							} else {
								current -= 2
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: "<",
						OnClicked: func() {
							if current-1 < 0 {
								current = len(segment) - 1
							} else {
								current--
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: "RESET",
						OnClicked: func() {
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[0]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: ">",
						OnClicked: func() {
							if current+1 > len(segment)-1 {
								current = 0
							} else {
								current++
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: ">>",
						OnClicked: func() {
							if current+2 > len(segment)-1 {
								current = 0
							} else {
								current += 2
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
					PushButton{
						Text: ">>>",
						OnClicked: func() {
							if current+3 > len(segment)-1 {
								current = 0
							} else {
								current += 3
							}
							b, err = walk.NewBitmapFromImageForDPI(getImage(segment[current]), 96)
							fatal(err)
							image.SetImage(b)
						},
					},
				},
			},
		},
	}.Run()
}

func getImage(seg segment) (pix image.Image) {
	req, err := getter("https://www.meteo.lv" + seg.image)
	fatal(err)
	img, err := png.Decode(bytes.NewReader(req))
	fatal(err)
	return img
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func getter(url string) (data []byte, err error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	for _, v := range cookies {
		req.AddCookie(v)
	}
	req.Header.Set("Referer", url)
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36")
	get, err := client.Do(req)
	if err != nil {
		return
	}
	defer get.Body.Close()
	if get.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", get.StatusCode)
	}
	data, err = ioutil.ReadAll(get.Body)
	if err != nil {
		return
	}
	return
}

func getCookie(url string) (cookie []*http.Cookie, err error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("Referer", url)
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return resp.Cookies(), nil
}
