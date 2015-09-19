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
	p.Width = 80
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
	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()
	ui.UseTheme("helloworld")
	s := createPar("Status", "", 0)
	i := createPar("Instructions", ":PRESS ESC TO QUIT", 2)
	c := createPar("Challenge", "", 4)
	o := createPar("Opponent", "", 6)
	w := createPar("Clock", "", 8)
	p := createPar("Points", "", 10)
	a := createPar("Actions", "", 12)
	d := createPar("Debug", "", 14)
	draw := func(msg gameMatchState) {
		s.Text = msg.Status
		c.Text = fmt.Sprintf("%v", msg.MyPlay)
		o.Text = fmt.Sprintf("%v", msg.OpponentPlay)
		w.Text = fmt.Sprintf("%v", msg.Clock)
		p.Text = fmt.Sprintf("%v", msg.Points)
		a.Text = fmt.Sprintf("%v", msg.Actions)
		ui.Render(s, i, c, o, w, p, a, d)
	}

	evt := ui.EventCh()
	var m = make(chan gameMatchState)

	go func() {
		var msg gameMatchState
		for {
			if err = websocket.JSON.Receive(ws, &msg); err != nil {
				msg.Status = "disconnected"
				d.Text = fmt.Sprintf("%v",err)
				draw(msg)
				break;
			}
			m <- msg
		}
	}()

	keyPressed := func(e ui.Event) keypress{
		if e.Key == ui.KeyBackspace2 {
			return keypress{
				Name: "down",
				KeyRune: rune(8),
				CharRune: rune(8),
			}

		}
		if e.Key == ui.KeySpace {
			return keypress{
				Name: "down",
				KeyRune: rune(0x20),
				CharRune: rune(0x20),
			}

		}
		return keypress{
			Name: "down",
			KeyRune: e.Ch,
			CharRune: e.Ch,
		}
	}
	for {
		select {
		case e := <-evt:
			d.Text = fmt.Sprintf("%v", e)
			if e.Type == ui.EventKey {
				if err := websocket.JSON.Send(ws,keyPressed(e)); err != nil {
					log.Fatal(err)
				}
			}

			if e.Type == ui.EventKey && e.Key == ui.KeyEsc {
				return
			}
		case msg := <-m:
			draw(msg)
		}
	}
}
