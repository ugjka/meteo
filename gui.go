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
	grid2, err := gtk.GridNew()
	fatal(err)
	grid2.SetBorderWidth(6)
	grid2.SetColumnHomogeneous(true)
	grid2.SetRowHomogeneous(true)
	grid2.SetColumnSpacing(6)
	grid.Attach(grid2, 0, 2, 1, 1)
	backward3, err := gtk.ButtonNewWithLabel("<<<")
	fatal(err)
	backward3.Connect("clicked", func() {
		if w.counter > 2 {
			w.counter -= 3
		} else {
			w.counter = len(w.seg) - w.counter - 3
		}
		w.update()
	})
	grid2.Attach(backward3, 0, 0, 1, 1)
	backward2, err := gtk.ButtonNewWithLabel("<<")
	fatal(err)
	backward2.Connect("clicked", func() {
		if w.counter > 1 {
			w.counter -= 2
		} else {
			w.counter = len(w.seg) - w.counter - 2
		}
		w.update()
	})
	grid2.Attach(backward2, 1, 0, 1, 1)
	backward1, err := gtk.ButtonNewWithLabel("<")
	backward1.Connect("clicked", func() {
		if w.counter > 0 {
			w.counter--
		} else {
			w.counter = len(w.seg) - w.counter - 1
		}
		w.update()
	})
	fatal(err)
	grid2.Attach(backward1, 2, 0, 1, 1)
	reset, err := gtk.ButtonNewWithLabel("reset")
	fatal(err)
	reset.Connect("clicked", func() {
		w.counter = 0
		w.update()
	})
	grid2.Attach(reset, 3, 0, 1, 1)
	forward1, err := gtk.ButtonNewWithLabel(">")
	fatal(err)
	forward1.Connect("clicked", func() {
		if w.counter < len(w.seg)-1 {
			w.counter++
			w.update()
		} else {
			w.counter = 0
			w.update()
		}
	})
	grid2.Attach(forward1, 4, 0, 1, 1)
	forward2, err := gtk.ButtonNewWithLabel(">>")
	fatal(err)
	forward2.Connect("clicked", func() {
		if w.counter+1 < len(w.seg)-1 {
			w.counter += 2
			w.update()
		} else {
			w.counter = 2 - (len(w.seg) - w.counter)
			w.update()
		}
	})
	grid2.Attach(forward2, 5, 0, 1, 1)
	forward3, err := gtk.ButtonNewWithLabel(">>>")
	fatal(err)
	forward3.Connect("clicked", func() {
		if w.counter+2 < len(w.seg)-1 {
			w.counter += 3
			w.update()
		} else {
			w.counter = 3 - (len(w.seg) - w.counter)
			w.update()
		}
	})
	grid2.Attach(forward3, 6, 0, 1, 1)
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
	req, err := getter("https://www.meteo.lv" + seg.image)
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
		return nil, fmt.Errorf("Status: %d", get.StatusCode)
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
