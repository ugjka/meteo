package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"regexp"

	_ "embed"

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

var modesStr = []string{
	"Nokrišņi",
	"Mākoņi",
	"Temperatūra",
	"Komforta temperatūra",
	"Vējš",
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"
const page = "https://www.meteo.lv"
const magicString = "?nid=557"

//go:embed logo.png
var icon []byte

func main() {
	iconDecoded, _ := png.Decode(bytes.NewBuffer(icon))
	iconBitmap, _ := walk.NewBitmapFromImageForDPI(iconDecoded, 96)

	imagebox := &walk.ImageView{}
	selector := &walk.ComboBox{}
	textedit := &walk.TextEdit{}
	mainWindow := &walk.MainWindow{}
	var forecast []string
	var err error
	var current = 0
	var mode = 0

	load := func(i int) {
		if len(forecast) == 0 {
			return
		}
		var bitmap = new(walk.Bitmap)
		rawimage, err := loadIMG(forecast[i])
		if err != nil {
			textedit.SetText(err.Error())
			textedit.SetVisible(true)
			imagebox.SetVisible(false)
		} else {
			textedit.SetVisible(false)
			imagebox.SetVisible(true)
			bitmap, _ = walk.NewBitmapFromImageForDPI(rawimage, 96)
			imagebox.SetImage(bitmap)
		}
	}

	MainWindow{
		AssignTo: &mainWindow,
		Title:    "Meteo.lv",
		Size: Size{
			Width:  650,
			Height: 400},
		Layout: VBox{},
		Children: []Widget{
			ComboBox{
				Model:        modesStr,
				CurrentIndex: 0,
				OnCurrentIndexChanged: func() {
					if selector.CurrentIndex() == mode {
						return
					}
					mode = selector.CurrentIndex()

					go func() {
						id := selector.CurrentIndex()
						forecast, err = loadForecast(modes[id])
						if err != nil {
							textedit.SetText(err.Error())
							textedit.SetVisible(true)
							imagebox.SetVisible(false)
						} else {
							load(current)
						}
					}()
				},
				AssignTo: &selector,
			},
			TextEdit{
				ReadOnly: true,
				Visible:  false,
				AssignTo: &textedit,
			},
			ImageView{
				AssignTo: &imagebox,
			},
			HSplitter{
				Children: []Widget{
					PushButton{
						Text: "<<<",
						OnClicked: func() {
							if current-2 < 0 {
								current = len(forecast) - 1
							} else {
								current -= 3
							}
							go load(current)
						},
					},
					PushButton{
						Text: "<<",
						OnClicked: func() {
							if current-2 < 0 {
								current = len(forecast) - 1
							} else {
								current -= 2
							}
							go load(current)
						},
					},
					PushButton{
						Text: "<",
						OnClicked: func() {
							if current-1 < 0 {
								current = len(forecast) - 1
							} else {
								current--
							}
							go load(current)
						},
					},
					PushButton{
						Text: "RESET",
						OnClicked: func() {
							current = 0
							go load(0)
						},
					},
					PushButton{
						Text: ">",
						OnClicked: func() {
							if current+1 > len(forecast)-1 {
								current = 0
							} else {
								current++
							}
							go load(current)
						},
					},
					PushButton{
						Text: ">>",
						OnClicked: func() {
							if current+2 > len(forecast)-1 {
								current = 0
							} else {
								current += 2
							}
							go load(current)
						},
					},
					PushButton{
						Text: ">>>",
						OnClicked: func() {
							if current+3 > len(forecast)-1 {
								current = 0
							} else {
								current += 3
							}
							go load(current)
						},
					},
				},
			},
		},
	}.Create()

	mainWindow.SetIcon(iconBitmap)

	cookies, err = loadCookies(page)
	if err != nil {
		textedit.SetText(err.Error())
		textedit.SetVisible(true)
		imagebox.SetVisible(false)
	}

	forecast, err = loadForecast(precip)
	if err != nil {
		textedit.SetText(err.Error())
		textedit.SetVisible(true)
		imagebox.SetVisible(false)
	} else {
		go load(0)
	}

	mainWindow.Run()
}

func loadIMG(i string) (image.Image, error) {
	req, err := get(page + i)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(req))
}

func get(url string) ([]byte, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, v := range cookies {
		req.AddCookie(v)
	}
	req.Header.Set("Referer", url)
	req.Header.Set("User-Agent", userAgent)
	get, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer get.Body.Close()
	if get.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", get.StatusCode)
	}
	return ioutil.ReadAll(get.Body)
}

func loadCookies(url string) ([]*http.Cookie, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", url)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Cookies(), nil
}

func loadForecast(url string) ([]string, error) {
	data, err := get(url + magicString)
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile(`([/]dynamic.*[.]png).*\d{2}[.]\d{2}[.]\d{4}.*\d{2}[:]\d{2}`)
	matches := reg.FindAllStringSubmatch(string(data), 100)
	images := make([]string, len(matches))
	for i, v := range matches {
		images[i] = v[1]
	}
	return images, nil
}
