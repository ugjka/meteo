package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"

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

//go:embed logo.png
var icon []byte

func main() {
	iconParsed, _ := png.Decode(bytes.NewBuffer(icon))
	iconBitmap, _ := walk.NewBitmapFromImageForDPI(iconParsed, 96)

	imagebox := &walk.ImageView{}
	selector := &walk.ComboBox{}
	textedit := &walk.TextEdit{}
	mainWindow := &walk.MainWindow{}
	var segs segments
	var err error
	var current = 0

	load := func(i int) {
		if len(segs) == 0 {
			return
		}
		var bitmap = new(walk.Bitmap)
		rawimage, err := loadIMG(segs[i])
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
		Title:    "Meteo",
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
					go func() {
						id := selector.CurrentIndex()
						segs, err = loadSegments(modes[id])
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
								current = len(segs) - 1
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
								current = len(segs) - 1
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
								current = len(segs) - 1
							} else {
								current--
							}
							go load(current)
						},
					},
					PushButton{
						Text: "RESET",
						OnClicked: func() {
							go load(0)
						},
					},
					PushButton{
						Text: ">",
						OnClicked: func() {
							if current+1 > len(segs)-1 {
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
							if current+2 > len(segs)-1 {
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
							if current+3 > len(segs)-1 {
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

	cookies, err = loadCookies("https://www.meteo.lv/")
	if err != nil {
		textedit.SetText(err.Error())
		textedit.SetVisible(true)
		imagebox.SetVisible(false)
	}

	segs, err = loadSegments(precip)
	if err != nil {
		textedit.SetText(err.Error())
		textedit.SetVisible(true)
		imagebox.SetVisible(false)
	} else {
		go load(0)
	}

	mainWindow.Run()
}

func loadIMG(seg segment) (image.Image, error) {
	req, err := get("https://www.meteo.lv" + seg.image)
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
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36")
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
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Cookies(), nil
}
