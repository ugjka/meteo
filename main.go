// meteo.lv vecā tipa kartes
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	_ "embed"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var cookies []*http.Cookie

const (
	precipitation = "https://www.meteo.lv/laiks/nokrisni/"
	clouds        = "https://www.meteo.lv/laiks/makoni/"
	temperature   = "https://www.meteo.lv/laiks/temperatura/"
	comfort       = "https://www.meteo.lv/laiks/komforta-temperatura/"
	wind          = "https://www.meteo.lv/laiks/vejs/"
)

var modes = []string{
	precipitation,
	clouds,
	temperature,
	comfort,
	wind,
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

const appErrorStr = `Radās sekojošā klūda: %s

Pārbaudiet savu interneta savienujumu un restartējiet aplikāciju.

Vai ziņojiet: esesmu@protonmail.com`

func main() {

	var imageView = new(walk.ImageView)
	var comboBox = new(walk.ComboBox)
	var textEdit = new(walk.TextEdit)
	var mainWindow = new(walk.MainWindow)
	var forecast []string
	var err error
	var current = 0
	var mode = 0

	var load = func(i int) {
		if len(forecast) == 0 {
			return
		}
		imageRaw, err := loadIMG(forecast[i])
		if err != nil {
			appError(textEdit, err)
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
							appError(textEdit, err)
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
				Mode:     ImageViewModeZoom,
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

	iconRaw, _ := png.Decode(bytes.NewBuffer(icon))
	iconBitmap, _ := walk.NewBitmapFromImageForDPI(iconRaw, 96)
	mainWindow.SetIcon(iconBitmap)

	cookies, err = loadCookies(page)
	if err == nil {
		forecast, err = loadForecast(precipitation)
	}
	if err != nil {
		appError(textEdit, err)
		textEdit.SetVisible(true)
		imageView.SetVisible(false)
	} else {
		go load(0)
	}

	mainWindow.Run()
}

func loadIMG(url string) (image.Image, error) {
	data, err := get(page + url)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
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
	matches := imageURLsReg.FindAllStringSubmatch(string(data), -1)
	imageURLs := make([]string, len(matches))
	for i, v := range matches {
		imageURLs[i] = v[1]
	}
	return imageURLs, nil
}

func appError(textEdit *walk.TextEdit, err error) {
	textEdit.SetText(fmt.Sprintf(strings.ReplaceAll(appErrorStr, "\n", "\r\n"), err))
}
