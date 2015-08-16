package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	ws "code.google.com/p/go.net/websocket"
	"log"

	"html/template"
)
type Octagon struct {
	first      *ws.Conn
	last       *ws.Conn
	disconnect chan struct{}
}

func NewOctagon() *Octagon {
	game := new(Octagon)
	game.disconnect = make(chan struct{})
	go func() {
		for {
			select {
			case <-game.disconnect:
				game.first.Close()
				game.last.Close()
				log.Printf("- game %s <-> %s disconnected", 
					game.first.RemoteAddr(), 
					game.last.RemoteAddr(), 
				)

			}
		}
	}()
	return game
}

func (g *Octagon) Clone() *Octagon {
	game := &Octagon{}
	game.first = g.first
	game.last = g.last
	game.disconnect = g.disconnect
	return game
}

func (g *Octagon) Connect(conn *ws.Conn) {
	if g.first != nil {
		log.Printf("- game last %s connected", conn.RemoteAddr())
		g.last = conn
	} else {
		log.Print("- game first %s connected", conn.RemoteAddr())
		g.first = conn
	}
}

func (g *Octagon) Disconnect() {
	var a struct{}
	g.disconnect <- a
}

func (g *Octagon) Broadcast(message string) {
	if ws.Message.Send(g.first, message) != nil {
		g.Disconnect()
	}
	if ws.Message.Send(g.last, message) != nil {
		g.Disconnect()
	}
}




type Player struct {
	conn *ws.Conn
	v *Valhalla
	name string
	password string //default idrinkyourmilkshake
	disconnect chan struct{}
}

var nametog int
func (p *Player) randomName(){
	if nametog == 0 {
		p.name = "Buttercup"
	} else {
		p.name = "Zelda"
	}
	nametog = 1 - nametog
}

func (p *Player) Receive(){
	var message string
	if ws.Message.Receive(p.conn, &message) != nil {
		var a struct{}
		p.disconnect <- a
	}
	p.Broadcast(message)
}
func (p *Player) Broadcast(message string){
	p.v.Broadcast(message)
}
func (p *Player) Listen() {
	for {
		select {
		case <-p.disconnect:
			return
		default:
			p.Receive()
		}
	}
}

func (v *Valhalla) Broadcast(message string){
	for p, _ := range v.allYall {
		ws.Message.Send(p.conn, message)
	}
}
// New gives you back a Player so fresh and clean clean, yo.
func (v *Valhalla) NewPlayer(conn *ws.Conn) *Player {
	player := &Player{
		v: v, 
		conn: conn, 
		disconnect: make(chan struct{}),
	}
	var a struct{}
	v.allYall[player] = a
	player.randomName()
	go func() {
		player.Listen()
		player.conn.Close()
		log.Printf("- player %s %s disconnected.", player.name, player.conn.RemoteAddr())
		delete(v.allYall, player)
	}()
	return player
}


type Valhalla struct {
	allYall map[*Player]struct{}
}

// Connect throws new players in the mix, yo!
func (v *Valhalla) Connect(conn *ws.Conn) *Player {
	log.Printf("- player connect: %s", conn)
	p := v.NewPlayer(conn)
	var a struct{}
	v.allYall[p] = a
	v.Broadcast("A new player has joined.")
	return p
}

func (v *Valhalla) Disconnect(p *Player){
	delete(v.allYall, p)
}


// var games GameList
var valhalla Valhalla
// store players in file
// allow players to choose opponents in a lobby
// open root -autoname> matchmaking -e> game start
//                                  -nothanks> lobby -chooseopponent> game start
//                                                   -chatupguests> chatter.

type echolog struct{
  sock *ws.Conn
}
func (e echolog) id() string{
	return e.sock.Request().RemoteAddr;
}
func (e echolog) receiveFail(){
  log.Printf("- %s couldn't receive.", e.id())
}
func (e echolog) connect(){
  log.Printf("- %s connected", e.id())
}
func (e echolog) sendFail(){
log.Printf("- %s couldn't send.", e.id())

}
func (e echolog) disconnected(){
log.Printf("- %s disconnected", e.id())
}
func (e echolog) message(msg error){
	log.Printf("- %s error %s", e.id(), msg);
}

type multiEcho struct{
	ws *ws.Conn
	log echolog
}
var multiEchoCons *map[*ws.Conn]*multiEcho;

func registerMultiEchoConn(sock *ws.Conn) *multiEcho {
	me := createMultiEchoConn(sock)
	(*multiEchoCons)[sock] = me
	return me
}
func createMultiEchoConn(sock *ws.Conn) *multiEcho {
	multi := multiEcho{
		ws: sock,
		log: echolog{sock: sock},
	};
	return &multi
}
func (m *multiEcho) Listen(){
	m.log.connect()
	var message string
	for {
		err := ws.Message.Receive(m.ws, &message)
		if err != nil {
			m.log.message(err)
			m.log.receiveFail()
			m.ws.Close()
			m.log.disconnected()
			delete(*multiEchoCons, m.ws)
			break;
		}
		for conn, me := range *multiEchoCons {
			err := ws.Message.Send(me.ws, message)
			if err != nil {
				me.log.message(err)
				me.log.sendFail()
				me.ws.Close()
				me.log.disconnected()
				delete(*multiEchoCons, conn)
			}
		}
	}
}

func multiEchoServer(sock *ws.Conn) {
	me := registerMultiEchoConn(sock);
	me.Listen();
}
func initMultiEcho(){
	mECons := make(map[*ws.Conn]*multiEcho);
	multiEchoCons = &mECons
}
type sharedBuffer struct {
	mECons map[*ws.Conn]*multiEcho
	buf bytes.Buffer
	write chan string
	conns chan *ws.Conn
	closes chan *ws.Conn
}
func (s *sharedBuffer) register(sock *ws.Conn){
	me := createMultiEchoConn(sock)
	s.mECons[sock] = me
	s.receive(me)
}
func (s *sharedBuffer) listen(){
	for {
		select {
		case message := <- s.write:
			s.integrate(message)
		case conn := <- s.conns:
			s.register(conn)
		case closeConn := <- s.closes:
			s.onClose(closeConn)
		}
	}
}
func (s *sharedBuffer) receive(m *multiEcho){
	m.log.connect()
	var message string
	for {
		err := ws.Message.Receive(m.ws, &message)
		if err != nil {
			m.log.message(err)
			m.log.receiveFail()
			m.ws.Close()
			m.log.disconnected()
			s.onClose(m.ws)
			break;
		}
		s.integrate(message)
	}
}
func (s *sharedBuffer) integrate(message string){
	s.buf.WriteString(message)
	s.broadcast()
}
func (s *sharedBuffer) onClose(closeConn *ws.Conn){
	delete(s.mECons, closeConn)
}
func (s *sharedBuffer) broadcast(){
	for _, me := range s.mECons {
		err := ws.Message.Send(me.ws, s.buf.String())
		if err != nil {
			me.log.message(err)
			me.log.sendFail()
			me.ws.Close()
			me.log.disconnected()
			s.onClose(me.ws)
		}
	}
}
func bufferServer(sock *ws.Conn){
	chatBuf.register(sock)
}
var chatBuf *sharedBuffer
func initBufferServer(){
	cB := sharedBuffer{mECons: make(map[*ws.Conn]*multiEcho)}
	chatBuf = &cB
	go chatBuf.listen()
}
func main() {
	initMultiEcho()
	initBufferServer()
	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD"))))

	http.Handle("/socket/multi_echo", ws.Handler(multiEchoServer))

	http.Handle("/socket/buffer", ws.Handler(bufferServer))
	http.Handle("/socket/new_game", ws.Handler(func(sock *ws.Conn) {
		player := valhalla.Connect(sock)
/*		log := echolog{sock: sock}
		var message string
		go func(){
			for {
				if ws.Message.Receive(sock, &message) != nil {
					log.receiveFail()
*/					var a struct{}

					player.disconnect <- a
/*				}
			}
			player.Broadcast(message)
		}()
*/
	}))

	template.New("things")
        fmt.Println("listening...", os.Getenv("PORT")) // Must be 5002 to work with frontend.
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
