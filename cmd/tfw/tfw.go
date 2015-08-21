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

type sharedBuffer struct {
	mECons map[*ws.Conn]*multiEcho
	buf    *bytes.Buffer
	write  chan string
	conns  chan *ws.Conn
	closes chan *ws.Conn
}

func (s *sharedBuffer) register(sock *ws.Conn) {
	me := createMultiEchoConn(sock)
	s.mECons[sock] = me
	s.receive(me)
}

type keypress struct {
	Name     string
	KeyRune  rune
	CharRune rune
}

func (s *sharedBuffer) receive(m *multiEcho) {
	m.log.connect()
	var message keypress
	for {
		err := ws.JSON.Receive(m.ws, &message)
		if err != nil {
			m.log.message(err)
			m.log.receiveFail()
			m.ws.Close()
			m.log.disconnected()
			s.onClose(m.ws)
			s.buf.WriteString("client disconnected")
			s.broadcast()
			break
		}
		m.log.got(message)
		s.integrate(message)
	}
}
func (s *sharedBuffer) integrate(kp keypress) {
	s.interpret(kp)
	s.broadcast()
}
func (s *sharedBuffer) interpret(kp keypress) {
	if kp.Name != "down" {
		return
	}
	if strconv.IsPrint(kp.KeyRune) {
		s.buf.WriteRune(kp.KeyRune)
		return
	}

	if kp.KeyRune == rune(8) { // backspace
		oldbuf := s.buf.Bytes()
		backOneChar := len(oldbuf) - 1
		s.buf = bytes.NewBuffer(oldbuf[:backOneChar])
	}

}
func (s *sharedBuffer) onClose(closeConn *ws.Conn) {
	delete(s.mECons, closeConn)
}
func (s *sharedBuffer) broadcast() {
	for _, me := range s.mECons {
		me.log.put(s.buf.String())
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
func bufferServer(sock *ws.Conn) {
	chatBuf.register(sock)
}

var chatBuf *sharedBuffer

func initBufferServer() {
	cB := sharedBuffer{mECons: make(map[*ws.Conn]*multiEcho), buf: bytes.NewBuffer([]byte{})}
	chatBuf = &cB
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
