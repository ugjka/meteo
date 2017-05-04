package main

import "github.com/gotk3/gotk3/gtk"

var icon = "./logo.png"

const (
	//foreca = "http://meteo.lv/laiks/"
	precip = "http://meteo.lv/laiks/nokrisni/"
	clouds = "http://meteo.lv/laiks/makoni/"
	temper = "http://meteo.lv/laiks/temperatura/"
	confor = "http://meteo.lv/laiks/komforta-temperatura/"
	windin = "http://meteo.lv/laiks/vejs/"
)

func main() {
	lietus := getSegments(precip)
	var win weather
	win.seg = lietus
	win.pixbuff = getPixbuf(lietus[0])
	gtk.Init(nil)
	win.initWidgets()
	gtk.Main()
}
