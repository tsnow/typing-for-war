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
	"strings"
	"sync"
	//	"time"
)

func init() {
	log.Println("starting typing-for-war...")
}

type visitor struct {
	sock *ws.Conn
}

func (v visitor) id() string {
	return v.sock.Request().RemoteAddr
}
func (v visitor) logReceiveFail() {
	log.Printf("- %s couldn't receive", v.id())
}
func (v visitor) logConnect() {
	log.Printf("- %s connected", v.id())
}
func (v visitor) logSendFail() {
	log.Printf("- %s couldn't send", v.id())

}
func (v visitor) logDisconnected() {
	log.Printf("- %s disconnected", v.id())
}
func (v visitor) logMessage(msg error) {
	log.Printf("- %s error %s", v.id(), msg)
}
func (v visitor) logGot(msg interface{}) {
	log.Printf("- %s <- \"%s\"", v.id(), msg)
}
func (v visitor) logPut(msg interface{}) {
	log.Printf("- %s -> \"%s\"", v.id(), msg)
}

type position string

const Fore position = "fore"
const Aft position = "aft"

type game struct {
	players map[position]*player
	gid gameID
}

func (p *player) id() string {
	return p.sock.Request().RemoteAddr
}
func (p *player) logReceiveFail() {
	log.Printf("- game %s - %s couldn't receive", p.g.gid, p.id())
}
func (p *player) logConnect() {
	log.Printf("- game %s - %s connected", p.g.gid, p.id())
}
func (p *player) logSendFail() {
	log.Printf("- game %s - %s couldn't send", p.g.gid, p.id())

}
func (p *player) logDisconnected() {
	log.Printf("- game %s - %s disconnected", p.g.gid, p.id())
}
func (p *player) logMessage(msg error) {
	log.Printf("- game %s - %s error %s", p.g.gid, p.id(), msg)
}
func (p *player) logGot(msg interface{}) {
	log.Printf("- game %s - %s <- \"%s\"", p.g.gid, p.id(), msg)
}
func (p *player) logPut(msg interface{}) {
	log.Printf("- game %s - %s -> \"%s\"", p.g.gid, p.id(), msg)
}

func newGame(gid gameID) *game {
	g := game{
		players: make(map[position]*player),
		gid: gid,
	}
	fore := player{
		pos: Fore,
		buf: bytes.NewBuffer([]byte{}),
		sock: nil,
		g: &g,
	}
	g.players[Fore] = &fore
	aft := player{
		pos: Aft,
		buf: bytes.NewBuffer([]byte{}),
		sock: nil,
		g: &g,
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
	return !(g.players[Fore].sock == nil) &&
		!(g.players[Aft].sock == nil)
}
func (v visitor) reject() {
	v.logConnect()
	err := ws.JSON.Send(v.sock, gameState{
		Status: NoGameAvailable,
	})
	if err != nil {
		v.logMessage(err)
		v.logSendFail()
	}
	v.logDisconnected()
	v.sock.Close()
}

type player struct {
	sock *ws.Conn
	pos position
	buf *bytes.Buffer
	g *game
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
		visitor{sock}.reject()
		return
	}
	
	var chosenPlayer *player
	for _, p := range g.players {
		if p.sock == nil {
			chosenPlayer = p
		}
	}
	chosenPlayer.sock = sock
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
	p.logConnect()

	var message keypress
	for {
		err := ws.JSON.Receive(p.sock, &message)
		if err != nil {
			p.logMessage(err)
			p.logReceiveFail()
			p.sock.Close()
			p.logDisconnected()
			p.onClose()
			g.broadcast()
			break
		}
		p.logGot(message)
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
		if len(oldbuf) == 0 {
			backOneChar=0
		}
		p.buf = bytes.NewBuffer(oldbuf[:backOneChar])
	}

}
func (p *player) onClose() {
	p.sock = nil
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
		if p.sock == nil {
			continue;
		}
		p.logPut(p.buf.String())
		err := ws.JSON.Send(p.sock, g.gameState(p))
		if err != nil {
			p.logMessage(err)
			p.logSendFail()
			p.sock.Close()
			p.logDisconnected()
			p.onClose()
		}
	}
}
func bufferServer(sock *ws.Conn) {
	g := parseGamePath(sock.Request().URL.Path)
	if g == nil {
		visitor{sock}.reject()
		return 
	}
	g.register(sock)
}
type gameID string
var games map[gameID]*game
func parseGamePath(url string) *game{
	gid := gameID(strings.TrimPrefix(url, gameRootPath()))
	return games[gid]
}
func buildGamePath(gid string) string{
	return gameRootPath() + gid
}
var mutex = &sync.Mutex{}
func createGame(name string){
	mutex.Lock()
	gid := gameID(name)
	games[gid] = newGame(gid)
	mutex.Unlock()
}
func gameRootPath() string{
	return "/game/"
}
func initBufferServer() {
	games = make(map[gameID]*game)
}
func releaseBufferServer() {
}

func main() {
	initBufferServer()
	createGame("sparklemotion")
	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD"))))

	http.Handle(gameRootPath(), ws.Handler(bufferServer))

	fmt.Println("listening...", os.Getenv("PORT")) // Must be 5002 to work with frontend.
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

}
