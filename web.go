package main

import (
	"bytes"
	"net/http"
	"os"

	ws "code.google.com/p/go.net/websocket"
	"log"

	"html/template"
)

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

type multiEcho struct {
	ws  *ws.Conn
	log echolog
}

var multiEchoCons *map[*ws.Conn]*multiEcho

func registerMultiEchoConn(sock *ws.Conn) *multiEcho {
	me := createMultiEchoConn(sock)
	(*multiEchoCons)[sock] = me
	return me
}
func createMultiEchoConn(sock *ws.Conn) *multiEcho {
	multi := multiEcho{
		ws:  sock,
		log: echolog{sock: sock},
	}
	return &multi
}
func (m *multiEcho) Listen() {
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
			break
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
	me := registerMultiEchoConn(sock)
	me.Listen()
}
func initMultiEcho() {
	mECons := make(map[*ws.Conn]*multiEcho)
	multiEchoCons = &mECons
}

type sharedBuffer struct {
	mECons map[*ws.Conn]*multiEcho
	buf    bytes.Buffer
	write  chan string
	conns  chan *ws.Conn
	closes chan *ws.Conn
}

func (s *sharedBuffer) register(sock *ws.Conn) {
	me := createMultiEchoConn(sock)
	s.mECons[sock] = me
	s.receive(me)
}
func (s *sharedBuffer) listen() {
	for {
		select {
		case message := <-s.write:
			s.integrate(message)
		case conn := <-s.conns:
			s.register(conn)
		case closeConn := <-s.closes:
			s.onClose(closeConn)
		}
	}
}
func (s *sharedBuffer) receive(m *multiEcho) {
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
			break
		}
		s.integrate(message)
	}
}
func (s *sharedBuffer) integrate(message string) {
	s.buf.WriteString(message)
	s.broadcast()
}
func (s *sharedBuffer) onClose(closeConn *ws.Conn) {
	delete(s.mECons, closeConn)
}
func (s *sharedBuffer) broadcast() {
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
func bufferServer(sock *ws.Conn) {
	chatBuf.register(sock)
}

var chatBuf *sharedBuffer

func initBufferServer() {
	cB := sharedBuffer{mECons: make(map[*ws.Conn]*multiEcho)}
	chatBuf = &cB
	go chatBuf.listen()
}
func main() {
}
