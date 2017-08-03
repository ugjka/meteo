package main

import (
	"net/http"

	"github.com/gotk3/gotk3/gtk"
)

var icon = "./logo.png"
var cookies []*http.Cookie

const (
	//foreca = "http://meteo.lv/laiks/"
	precip = "https://www.meteo.lv/laiks/nokrisni/"
	clouds = "https://www.meteo.lv/laiks/makoni/"
	temper = "https://www.meteo.lv/laiks/temperatura/"
	confor = "https://www.meteo.lv/laiks/komforta-temperatura/"
	windin = "https://www.meteo.lv/laiks/vejs/"
)

func main() {
	var err error
	cookies, err = getCookie("https://www.meteo.lv/")
	if err != nil {
		panic(err)
	}
	lietus := getSegments(precip)
	var win weather
	win.seg = lietus
	win.pixbuff = getPixbuf(lietus[0])
	gtk.Init(nil)
	win.initWidgets()
	gtk.Main()
}
