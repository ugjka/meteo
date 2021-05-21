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

var modesDisplay = []string{
	"Nokrišņi",
	"Mākoņi",
	"Temperatūra",
	"Komforta temperatūra",
	"Vējš",
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"
const page = "https://www.meteo.lv"
const magicString = "?nid=557"

var imageURLsReg = regexp.MustCompile(`([/]dynamic.*[.]png).*\d{2}[.]\d{2}[.]\d{4}.*\d{2}[:]\d{2}`)

//go:embed logo.png
var icon []byte

func main() {
	iconRaw, _ := png.Decode(bytes.NewBuffer(icon))
	iconBitmap, _ := walk.NewBitmapFromImageForDPI(iconRaw, 96)

	imageView := &walk.ImageView{}
	comboBox := &walk.ComboBox{}
	textEdit := &walk.TextEdit{}
	mainWindow := &walk.MainWindow{}
	var forecast []string
	var err error
	var current = 0
	var mode = 0

	load := func(i int) {
		if len(forecast) == 0 {
			return
		}
		imageRaw, err := loadIMG(forecast[i])
		if err != nil {
			textEdit.SetText(err.Error())
			textEdit.SetVisible(true)
			imageView.SetVisible(false)
		} else {
			textEdit.SetVisible(false)
			imageView.SetVisible(true)
			imageBitmap, _ := walk.NewBitmapFromImageForDPI(imageRaw, 96)
			imageView.SetImage(imageBitmap)
		}
	}

	MainWindow{
		AssignTo: &mainWindow,
		Title:    "Meteo.lv",
		Size: Size{
			Width:  652,
			Height: 553},
		Layout: VBox{},
		Children: []Widget{
			ComboBox{
				Model:        modesDisplay,
				CurrentIndex: 0,
				OnCurrentIndexChanged: func() {
					if comboBox.CurrentIndex() == mode {
						return
					}
					mode = comboBox.CurrentIndex()

					go func() {
						forecast, err = loadForecast(modes[mode])
						if err != nil {
							textEdit.SetText(err.Error())
							textEdit.SetVisible(true)
							imageView.SetVisible(false)
						} else {
							load(current)
						}
					}()
				},
				AssignTo: &comboBox,
			},
			TextEdit{
				ReadOnly: true,
				Visible:  false,
				AssignTo: &textEdit,
			},
			ImageView{
				AssignTo: &imageView,
			},
			Composite{
				Layout: HBox{},
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
		textEdit.SetText(err.Error())
		textEdit.SetVisible(true)
		imageView.SetVisible(false)
	}

	forecast, err = loadForecast(precip)
	if err != nil {
		textEdit.SetText(err.Error())
		textEdit.SetVisible(true)
		imageView.SetVisible(false)
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

func get(URL string) ([]byte, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	for _, v := range cookies {
		req.AddCookie(v)
	}
	req.Header.Set("Referer", URL)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func loadCookies(URL string) ([]*http.Cookie, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", URL)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Cookies(), nil
}

func loadForecast(URL string) ([]string, error) {
	data, err := get(URL + magicString)
	if err != nil {
		return nil, err
	}
	matches := imageURLsReg.FindAllStringSubmatch(string(data), -1)
	imageURLs := make([]string, len(matches))
	for i, v := range matches {
		imageURLs[i] = v[1]
	}
	return imageURLs, nil
}
