package main

import (
	"bytes"
	ws "code.google.com/p/go.net/websocket"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var serverAddr string
var once sync.Once

func startServer() {
	http.Handle(gameRootPath(), ws.Handler(bufferServer))
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
}

func newConfig(t *testing.T, path string) *ws.Config {
	config, _ := ws.NewConfig(fmt.Sprintf("ws://%s%s", serverAddr, path), "http://localhost")
	return config
}

func createClient(t *testing.T, resource string) *ws.Conn {

	// websocket.Dial()
	client, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	conn, err := ws.NewClient(newConfig(t, resource), client)
	if err != nil {
		t.Errorf("WebSocket handshake error: %v", err)
		return nil
	}
	return conn
}
func verifyReceive(t *testing.T, gid string, conn *ws.Conn, msg []byte) {
	var actual_msg = make([]byte, 512)

	n, err := conn.Read(actual_msg)
	if err != nil {
		t.Errorf("Read: %s %v",gid, err)
	}
	actual_msg = actual_msg[0:n]
	if !bytes.Equal(msg, actual_msg) {
		t.Errorf("Echo: %s expected %q got %q", gid, msg, actual_msg)
	}

}

func TestGameBackspace(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g := nextGame()
	conn1 := createClient(t, buildGamePath(g))
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, buildGamePath(g))
	if conn2 == nil {
		return
	}

	bkspmsg := []byte("{\"Name\":\"down\",\"KeyRune\":8}")
	hmsg := []byte("{\"Name\":\"down\",\"KeyRune\":72}")
	imsg := []byte("{\"Name\":\"down\",\"KeyRune\":73}")
	h := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"\",\"MyPlay\":\"H\"}")
	hi := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"\",\"MyPlay\":\"HI\"}")
	if _, err := conn1.Write(hmsg); err != nil {
		t.Errorf("Write: %v", err)
	}

	verifyReceive(t, g, conn1, h)

	if _, err := conn1.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, g, conn1, hi)

	if _, err := conn1.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, g, conn1, h)
	conn1.Close()
	conn2.Close()
	releaseBufferServer()
}

func TestGameReconnectConn(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g := nextGame()
	conn1 := createClient(t, buildGamePath(g))
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, buildGamePath(g))
	if conn2 == nil {
		return
	}

	bkspmsg := []byte("{\"Name\":\"down\",\"KeyRune\":8}")
	hmsg := []byte("{\"Name\":\"down\",\"KeyRune\":72}")
	imsg := []byte("{\"Name\":\"down\",\"KeyRune\":73}")
	h1 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"\",\"MyPlay\":\"H\"}")
	h2 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	i1 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"I\",\"MyPlay\":\"H\"}")
	i2 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"H\",\"MyPlay\":\"I\"}")
	if _, err := conn1.Write(hmsg); err != nil {
		t.Errorf("Write: %v", err)
	}

	verifyReceive(t, g, conn1, h1)
	verifyReceive(t, g, conn2, h2)

	if _, err := conn2.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, g, conn1, i1)
	verifyReceive(t, g, conn2, i2)

	conn1.Close()

	if _, err := conn2.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	que := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	verifyReceive(t, g, conn2, que)

	wait2 := []byte("{\"Status\":\"waiting_for_opponent\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	if _, err := conn2.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, g, conn2, wait2)

	conn1 = createClient(t, "/buffer")
	if conn1 == nil {
		return
	}
	if _, err := conn1.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	wait1 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"\",\"MyPlay\":\"HI\"}")
	regame2 := []byte("{\"Status\":\"waiting_for_opponent\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	verifyReceive(t, g, conn1, wait1)
	verifyReceive(t, g, conn2, regame2)
	conn2.Close()
	conn1.Close()
	releaseBufferServer()
}
