package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type weather struct {
	window   *gtk.Window
	label    *gtk.Label
	pixbuff  *gdk.Pixbuf
	image    *gtk.Image
	combo    *gtk.ComboBoxText
	eventbox *gtk.EventBox
	seg      segments
	counter  int
	onOpen   func()
}

func (w *weather) initWidgets() {
	var err error
	w.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	fatal(err)
	w.window.SetTitle("Meteo.lv")
	w.window.SetPosition(gtk.WIN_POS_CENTER)
	w.window.SetSizeRequest(640, 320)
	w.window.SetBorderWidth(6)
	err = w.window.SetIconFromFile(icon)
	fatal(err)
	_, err = w.window.Connect("destroy", gtk.MainQuit)
	fatal(err)
	grid, err := gtk.GridNew()
	fatal(err)
	grid.SetColumnHomogeneous(true)
	grid.SetColumnSpacing(6)
	grid.SetRowSpacing(6)
	w.window.Add(grid)
	w.combo, err = gtk.ComboBoxTextNew()
	fatal(err)
	w.combo.Append("1", "Lietus")
	w.combo.Append("2", "Mākoņi")
	w.combo.Append("3", "Temperatūra")
	w.combo.Append("4", "Komf. Temperatūra")
	w.combo.Append("5", "Vējš")
	w.combo.SetActiveID("1")
	w.combo.Connect("changed", w.comboChange)
	grid.Attach(w.combo, 0, 0, 1, 1)
	w.image, err = gtk.ImageNewFromPixbuf(w.pixbuff)
	w.eventbox, err = gtk.EventBoxNew()
	fatal(err)
	w.eventbox.Add(w.image)
	_, err = w.eventbox.Connect("button-press-event", func(b *gtk.EventBox, e *gdk.Event) {
		if gdk.EventKeyNewFromEvent(e).Type() == gdk.EVENT_BUTTON_PRESS && gdk.EventButtonNewFromEvent(e).ButtonVal() == uint(3) {
			if w.counter > 0 {
				w.counter--
			} else {
				w.counter = len(w.seg) - 1
			}
		}
		if gdk.EventKeyNewFromEvent(e).Type() == gdk.EVENT_BUTTON_PRESS && gdk.EventButtonNewFromEvent(e).ButtonVal() == uint(1) {
			if w.counter < (len(w.seg) - 1) {
				w.counter++
			} else {
				w.counter = 0
			}
		}
		w.update()
	})
	fatal(err)
	grid.Attach(w.eventbox, 0, 1, 1, 1)
	w.window.ShowAll()
}

func (w *weather) comboChange() {
	switch w.combo.GetActiveID() {
	case "1":
		w.seg = getSegments(precip)
	case "2":
		w.seg = getSegments(clouds)
	case "3":
		w.seg = getSegments(temper)
	case "4":
		w.seg = getSegments(confor)
	case "5":
		w.seg = getSegments(windin)
	}
	w.update()
}

func (w *weather) update() {
	w.pixbuff = getPixbuf(w.seg[w.counter])
	w.image.SetFromPixbuf(w.pixbuff)
}

func getPixbuf(seg segment) (pix *gdk.Pixbuf) {
	req, err := getter("http://meteo.lv" + seg.image)
	fatal(err)
	img, err := png.Decode(bytes.NewReader(req))
	fatal(err)
	rect := img.Bounds()
	loader, err := gdk.PixbufLoaderNew()
	fatal(err)
	defer loader.Close()
	loader.SetSize(rect.Dx(), rect.Dy())
	_, err = io.Copy(loader, bytes.NewReader(req))
	fatal(err)
	pix, err = loader.GetPixbuf()
	fatal(err)
	return
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func getter(url string) (data []byte, err error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Location", "http://meteo.lv/laiks/")
	get, err := client.Do(req)
	if err != nil {
		return
	}
	defer get.Body.Close()
	if get.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status: %d", get.StatusCode)
	}
	data, err = ioutil.ReadAll(get.Body)
	if err != nil {
		return
	}
	return
}
