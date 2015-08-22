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
	g := game{players: make(map[position]*player)}
	fore := player{
		pos: Fore,
		buf: bytes.NewBuffer([]byte{}),
		me: nil,
	}
	g.players[Fore] = &fore
	aft := player{
		pos: Aft,
		buf: bytes.NewBuffer([]byte{}),
		me: nil,
	}
	g.players[Aft] = &aft
	return &g
}
func (g *game) fore() *player {
	return g.players[Fore]
}
func (g *game) aft() *player {
	return g.players[Aft]
}
func (g *game) gameFull() bool {
	return !(g.players[Fore].me == nil) &&
		!(g.players[Aft].me == nil)
}
func (g *game) rejectVisitor(sock *ws.Conn) {
	me := createMultiEchoConn(sock)
	//TODO: this is an example of a site visitor behavior.
	me.log.connect()
	err := ws.JSON.Send(me.ws, gameState{
		Status: NoGameAvailable,
	})
	if err != nil {
		me.log.message(err)
		me.log.sendFail()
	}
	me.log.disconnected()
	me.ws.Close()
}

type player struct {
	me  *multiEcho
	pos position
	buf *bytes.Buffer
}

func (g *game) getPlayer(pos position) *player {
	return g.players[pos]
}
func (g *game) otherPlayer(pos position) *player {
	var out position
	if pos == Fore {
		out = Aft
	} else if pos == Aft {
		out = Fore
	}
	return g.getPlayer(out)
}
func (g *game) register(sock *ws.Conn) {
	if g.gameFull() {
		g.rejectVisitor(sock)
		return
	}
	
	var chosenPlayer *player
	for _, p := range g.players {
		if p.me == nil {
			chosenPlayer = p
		}
	}
	me := createMultiEchoConn(sock)
	chosenPlayer.me = me
	g.receive(chosenPlayer)
}

type keypress struct {
	Name     string
	KeyRune  rune
	CharRune rune
}

type status string
const WaitingForOpponent status = "waiting_for_opponent"
const NoGameAvailable status = "no_games_available"
const Gaming status = "gaming"

type gameState struct {
	Status status
	OpponentPlay string
	MyPlay string
}

func (g *game) receive(p *player) {
	//TODO: make multiEcho into player, including position, add to logging
	p.me.log.connect()

	var message keypress
	for {
		err := ws.JSON.Receive(p.me.ws, &message)
		if err != nil {
			p.me.log.message(err)
			p.me.log.receiveFail()
			p.me.ws.Close()
			p.me.log.disconnected()
			p.onClose()
			g.broadcast()
			break
		}
		p.me.log.got(message)
		g.integrate(p, message)
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
func (p *player) onClose() {
	p.me = nil
}
func (g *game) gameState(p *player) gameState{
	o := g.otherPlayer(p.pos)
	var state status
	if g.gameFull(){
		state = Gaming
	} else {
		state = WaitingForOpponent
	}
	return gameState{
		Status: state,
		OpponentPlay: o.buf.String(),
		MyPlay: p.buf.String(),
	}
}
func (g *game) broadcast() {
	for _, p := range g.players {
		if p.me == nil {
			continue;
		}
		p.me.log.put(p.buf.String())
		err := ws.JSON.Send(p.me.ws, g.gameState(p))
		if err != nil {
			p.me.log.message(err)
			p.me.log.sendFail()
			p.me.ws.Close()
			p.me.log.disconnected()
			p.onClose()
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
