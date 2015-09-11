package main

import (
	"bytes"
	ws "golang.org/x/net/websocket"
	"fmt"
	//	"github.com/tsnow/typing-for-war/engine"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
func (v visitor) logNoGameAvailable() {
	log.Printf("- %s no_game_available", v.id())
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
	players   map[position]*player
	gid       gameID
	objective string // TODO: Create objective struct and implement and test operations against it. (compare, partition_play{correct,wrong,left}, completed, list_of_errors.)
	clock     int
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

func newGame(gid gameID, objective string) *game {
	g := game{
		players:   make(map[position]*player),
		gid:       gid,
		objective: objective,
	}
	fore := player{
		pos:     Fore,
		sock:    nil,
		g:       &g,
	}
	g.players[Fore] = &fore
	aft := player{
		pos:     Aft,
		sock:    nil,
		g:       &g,
	}
	g.players[Aft] = &aft
	g.resetGame()
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
	v.logNoGameAvailable()
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
	sock    *ws.Conn
	pos     position
	buf     *bytes.Buffer
	g       *game
	points  int
	endTime int
	playerName string
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
	g.broadcast()
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
const GameStarting status = "game_starting"
const GameOver status = "game_over"
const GameWon status = "game_won"
const GameLost status = "game_lost"

type playState [3]string

type gameState struct {
	Status       status
	OpponentPlay playState
	MyPlay       playState
	Objective    string
	Clock        int
	Points       int
}

func (g *game) receive(p *player) {
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
		g.broadcast()
	}
}
func (g *game) integrate(p *player, kp keypress) {
	state := g.gameStatus(p)
	if state != Gaming {
		return
	}
	if completedGame(g.objective, p.buf.String()) {
		// here there be attacks
		o := g.otherPlayer(p.pos)
		g.interpret(o, kp)
	} else {
		g.interpret(p, kp)
	}
	if p.endTime < 0 && completedGame(g.objective, p.buf.String()) {
		g.distributePoints(p)
	}
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
			backOneChar = 0
		}
		p.buf = bytes.NewBuffer(oldbuf[:backOneChar])
	}

}
func (p *player) onClose() {
	p.sock = nil
}
func (g *game) gameStatus(p *player) status {
	o := g.otherPlayer(p.pos)
	var state status
	if p.endTime >= 0 && o.endTime >= 0 {
		if p.points > o.points {
			state = GameWon
		} else {
			state = GameLost
		}
		return state
	}
	if g.gameFull() && g.clock < 0 {
		state = GameStarting
	} else if g.gameFull() && g.clock > 0 {
		state = Gaming
	} else if g.gameFull() && g.clock == 0 {
		state = GameOver
	} else {
		state = WaitingForOpponent
	}
	return state
}
func (g *game) gameState(p *player) gameState {
	o := g.otherPlayer(p.pos)
	state := g.gameStatus(p)
	switch state {
	case GameStarting, WaitingForOpponent:
		return gameState{
			Status: state,
			Clock:        g.clock,
			Points:       p.points,
		}
	}
	return gameState{
		Status:       state,
		OpponentPlay: GoodBadLeft(g.objective, o.buf.String()),
		MyPlay:       GoodBadLeft(g.objective, p.buf.String()),
		Objective:    g.objective,
		Clock:        g.clock,
		Points:       p.points,
	}
}
func (g *game) distributePoints(p *player) {
	p.endTime = g.clock
	p.points = p.points + calcPoints(g.objective, p.buf.String()) + p.endTime
}
func calcPoints(objective string, attempt string) int {
	gbl := GoodBadLeft(objective, attempt)
	good := gbl[0]
	bad := gbl[1]
	out := len(good) - len(bad)
	if out <= 0 {
		return 0
	}
	return out
}
func completedGame(objective string, attempt string) bool {
	gbl := GoodBadLeft(objective, attempt)
	bad := gbl[1]
	left := gbl[2]
	if len(bad)+len(left) == 0 {
		return true
	}
	return false
}
func (g *game) goClock() bool{
	if !g.gameFull() {
		return true // pause
	}
	if g.clock == 0 { //not gaming
		return true //Shouldn't be reachable
	} else if g.clock == 1 { //game done
		fore := g.players[Fore]
		if fore.endTime < 0 {
			g.distributePoints(fore)
		}
		aft := g.players[Aft]
		if aft.endTime < 0 {
			g.distributePoints(aft)
		}
		g.resetGame()
		return false
	} else if g.clock > 0 {
		g.clock = g.clock - 1
		return false
	} else if g.clock == -1 { // time to start
		v, _ := g.gameSettings()
		g.clock = v.clock
		return false
	} else { // countdown to start
		g.clock = g.clock + 1
		return false
	}
}
func (g *game) tick() {
	if g.goClock() {
		return
	}
	g.broadcast()
}

func (g *game) gameSettings() (game, game){
	var o game
	z := len(gameSettings) - 1
	v := len(gameSettings)
	for i, h := range gameSettings {
		if g.objective == h.objective {
			o = h
			v = i + 1
		}
	}
	if v >= z {
		return o, gameSettings[0]
	}
	return o, gameSettings[v]
}
func (g *game) resetGame(){
	_, z := g.gameSettings()
	g.objective = z.objective
	g.clock = -10
	g.players[Fore].buf = bytes.NewBuffer([]byte{})
	g.players[Aft].buf = bytes.NewBuffer([]byte{})
	g.players[Fore].endTime = -1
	g.players[Aft].endTime = -1
}

func GoodBadLeft(objective string, attempt string) playState {
	good := bytes.Buffer{}
	bad := bytes.Buffer{}
	left := bytes.Buffer{}
	furthest := -1
	for i := 0; i < len(attempt); i++ {
		furthest = i
		if i == len(objective) { // e.g. we've gone beyond the objective characters
			bad.WriteString(attempt[i:])
			break
		}
		if attempt[i] != objective[i] {
			bad.WriteString(attempt[i:])
			left.WriteString(objective[i:])
			break
		}
		good.WriteByte(attempt[i])
	}
	if bad.Len() == 0 && left.Len() == 0 && (furthest+1) < len(objective) {
		left.WriteString(objective[(furthest + 1):])
	}
	return playState{good.String(), bad.String(), left.String()}
}
func (g *game) broadcast() {
	for _, p := range g.players {
		if p.sock == nil {
			continue
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
	log.Printf("%s %s", sock.Request().Method, sock.Request().URL.Path)
	g := parseGamePath(sock.Request().URL.Path)
	if g == nil {
		visitor{sock}.reject()
		return
	}
	g.register(sock)
}

type gameID string

var games map[gameID]*game

func parseGamePath(url string) *game {
	gid := gameID(strings.TrimPrefix(url, gameRootPath()))
	return games[gid]
}
func buildGamePath(gid string) string {
	return gameRootPath() + gid
}

var mutex = &sync.Mutex{}

func createGame(name string, objective string) {
	mutex.Lock()
	gid := gameID(name)
	games[gid] = newGame(gid, objective)
	mutex.Unlock()
}
func gameRootPath() string {
	return "/game/"
}
func initBufferServer() {
	games = make(map[gameID]*game)
	go func() {
		c := time.Tick(time.Second)
		for _ = range c {
			mutex.Lock()
			for _, game := range games {
				game.tick()
			}
			mutex.Unlock()
		}
	}()
}
var gameSettings []game = []game{
	game{
		objective: "CRY HAVOK AND LET SLIP THE DOGS OF WAR",
		clock: 10,
	},
	game{
		objective: "FLORETED CHOREA ANAGRAMMATICALLY LOCULATION REPREDICT",
		clock: 15,
	},
	game{
		objective: "TEH",
		clock: 2,
	},
	game{
		objective: "WINRAR",
		clock: 3,
	},
	game{
		objective: "CRY HAVOK N LET SLIP THE GODS OF WART",
		clock: 5,
	},
	game{
		objective: "WHY AM I UNREACHABLE????",
		clock: 5,
	},
}
func releaseBufferServer() {
}

func main() {
	initBufferServer()
	createGame("sparklemotion", "CRY HAVOK AND LET SLIP THE DOGS OF WAR")
	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD"))))

	http.Handle(gameRootPath(), ws.Handler(bufferServer))

	fmt.Println("listening...", os.Getenv("PORT")) 
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}

}
