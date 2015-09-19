package main

import (
	ui "github.com/gizak/termui"
	//	"bytes"
	"fmt"
	websocket "golang.org/x/net/websocket"
	//	"github.com/tsnow/typing-for-war/engine"
	"log"
	//	"net/http"
	//	"os"
	//	"strconv"
	//	"strings"
	//	"sync"
	//	"time"
)

func init() {
	log.Println("starting typing-for-war...")
}

type gameMatchState struct {
	Status       string
	OpponentPlay [3]string
	MyPlay       [3]string
	Objective    string
	Clock        int
	Points       int
	Actions      []string
}

type keypress struct {
	Name     string
	KeyRune  rune
	CharRune rune
}

func createPar(name string, value string, y int) *ui.Par {
	p := ui.NewPar(value)
	p.Height = 3
	p.Width = 50
	p.Y = y
	p.TextFgColor = ui.ColorWhite
	p.Border.Label = name
	p.Border.FgColor = ui.ColorCyan
	return p
}
func main() {
	origin := "http://localhost:5002/"
	url := "ws://localhost:5002/game/sparklemotion"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := ws.Write([]byte("hello, world!\n")); err != nil {
		log.Fatal(err)
	}
	var m = make(chan gameMatchState)

	go func() {
		var msg = gameMatchState{}
		for {
			if err = websocket.JSON.Receive(ws, &msg); err != nil {
				log.Fatal(err)
			}
			m <- msg
		}
	}()

	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()
	ui.UseTheme("helloworld")
	s := createPar("Status", "", 0)
	i := createPar("Instructions", ":PRESS q TO QUIT DEMO", 2)
	c := createPar("Challenge", "", 4)
	o := createPar("Opponent", "", 6)
	w := createPar("Clock", "", 8)
	p := createPar("Points", "", 10)
	a := createPar("Actions", "", 12)
	draw := func(msg gameMatchState) {
		s.Text = msg.Status
		c.Text = "[]"
		o.Text = "[]"
		w.Text = fmt.Sprintf("%v", msg.Clock)
		p.Text = fmt.Sprintf("%v", msg.Points)
		a.Text = fmt.Sprintf("%v", msg.Actions)
		ui.Render(s, i, c, o, w, p, a)
	}

	evt := ui.EventCh()

	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey && e.Ch == 'q' {
				return
			}
		case msg := <-m:
			draw(msg)
		}
	}
	/*
		ui.Body.AddRows(
			ui.NewRow(
				ui.NewCol(6, 0, widget0),
				ui.NewCol(6, 0, widget1)),
			ui.NewRow(
				ui.NewCol(3, 0, widget2),
				ui.NewCol(3, 0, widget30, widget31, widget32),
				ui.NewCol(6, 0, widget4)))

		// calculate layout
		ui.Body.Align()

		ui.Render(ui.Body)
	*/

}
