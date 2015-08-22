package main

import (
	"bytes"
	ws "code.google.com/p/go.net/websocket"
	"fmt"
	//	"github.com/tsnow/typing-for-war/engine"
	"log"
	"net/http"
	"os"
	"strconv"
	//	"time"
)

func init() {
	log.Println("starting typing-for-war...")
}

type echolog struct {
	sock *ws.Conn
}

func (e echolog) id() string {
	return e.sock.Request().RemoteAddr
}
func (e echolog) receiveFail() {
	log.Printf("- %s couldn't receive.", e.id())
}
func (e echolog) connect() {
	log.Printf("- %s connected", e.id())
}
func (e echolog) sendFail() {
	log.Printf("- %s couldn't send.", e.id())

}
func (e echolog) disconnected() {
	log.Printf("- %s disconnected", e.id())
}
func (e echolog) message(msg error) {
	log.Printf("- %s error %s", e.id(), msg)
}
func (e echolog) got(msg interface{}) {
	log.Printf("- %s <- \"%s\"", e.id(), msg)
}
func (e echolog) put(msg interface{}) {
	log.Printf("- %s -> \"%s\"", e.id(), msg)
}

type multiEcho struct {
	ws  *ws.Conn
	log echolog
}

func createMultiEchoConn(sock *ws.Conn) *multiEcho {
	multi := multiEcho{
		ws:  sock,
		log: echolog{sock: sock},
	}
	return &multi
}

type position string

const Fore position = "fore"
const Aft position = "aft"

type game struct {
	players map[position]*player
}

func newGame() *game {
	g := game{players: make(map[*ws.Conn]*player)}
	return &g
}
func (g *game) fore() *multiEcho {
	return g.players[Fore]
}
func (g *game) aft() *multiEcho {
	return g.players[Aft]
}
func (g *game) gameFull() bool {
	return !(g.players[Fore] == nil) &&
		!(g.players[Aft] == nil)
}
func (g *game) rejectVisitor(sock *ws.Conn) {
	me := createMultiEchoConn(sock)
	//TODO: this is an example of a "visitor" behavior.
	me.log.connect()
	err := ws.Message.Send(me.ws, "game full")
	if err != nil {
		me.log.message(err)
		me.log.sendFail()
	}
	m.log.disconnected()
	me.ws.Close()
}

type player struct {
	me  *multiEcho
	pos position
	buf *bytes.Buffer
}

func (g *game) myPlayer(pos position) *player {
	return g.players[pos]
}
func (g *game) otherPlayer(pos position) *player {
	if pos == Fore {
		return g.myPlayer(Aft)
	} else if pos == Aft {
		return g.myPlayer(Fore)
	}
}
func (g *game) register(sock *ws.Conn) {
	if g.gameFull() {
		g.rejectVisitor(sock)
		return
	}

	var pos position
	if g.players[Fore] == nil {
		pos = Fore
	} else if g.players[Aft] == nil {
		pos = Aft
	}
	me := createMultiEchoConn(sock)
	p := player{
		pos: pos,
		buf: bytes.NewBuffer([]byte{}),
		me:  me,
	}
	g.players[pos] = &p
	g.receive(p)
}

type keypress struct {
	Name     string
	KeyRune  rune
	CharRune rune
}

func (g *game) receive(p *player) {
	//TODO: make multiEcho into player, including position, add to logging
	o := g.otherPlayer(p.pos)
	p.me.log.connect()

	var message keypress
	for {
		err := ws.JSON.Receive(p.me.ws, &message)
		if err != nil {
			p.me.log.message(err)
			p.me.log.receiveFail()
			p.me.ws.Close()
			p.me.log.disconnected()
			g.onClose(p.pos)
			p.buf.WriteString("was disconnected")
			o.buf.WriteString("opponent disconnected")
			g.broadcast()
			break
		}
		p.me.log.got(message)
		g.integrate(message)
	}
}
func (g *game) integrate(p *player, kp keypress) {
	g.interpret(p, kp)
	g.broadcast()
}
func (g *game) interpret(p *player, kp keypress) {
	if kp.Name != "down" {
		return
	}
	if strconv.IsPrint(kp.KeyRune) {
		p.buf.WriteRune(kp.KeyRune)
		return
	}

	if kp.KeyRune == rune(8) { // backspace
		oldbuf := p.buf.Bytes()
		backOneChar := len(oldbuf) - 1
		p.buf = bytes.NewBuffer(oldbuf[:backOneChar])
	}

}
func (g *game) onClose(pos position) {
	delete(g.players, pos)
}
func (g *game) broadcast() {
	for _, p := range g.players {
		p.me.log.put(p.buf.String())
		err := ws.Message.Send(p.me.ws, p.buf.String())
		if err != nil {
			p.me.log.message(err)
			p.me.log.sendFail()
			p.me.ws.Close()
			p.me.log.disconnected()
			g.onClose(p.pos)
		}
	}
}
func bufferServer(sock *ws.Conn) {
	persistGame.register(sock)
}

var persistGame *game

func initBufferServer() {
	g := newGame()
	persistGame = g
}
func main() {
	initBufferServer()
	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD"))))

	http.Handle("/socket/buffer", ws.Handler(bufferServer))

	fmt.Println("listening...", os.Getenv("PORT")) // Must be 5002 to work with frontend.
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

}
