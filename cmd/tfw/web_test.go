package main

import (
	"bytes"
	ws "code.google.com/p/go.net/websocket"
	"encoding/json"
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
		t.Errorf("Read: %s %v", gid, err)
	}
	actual_msg = actual_msg[0:n]
	if !bytes.Equal(msg, actual_msg) {
		t.Errorf("Echo: %s expected \n%q\n got \n%q", gid, msg, actual_msg)
	}
}
func dontCreateGame(gid string) {
}
func TestGameDoesntExist(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g := "doesnt_exist"
	dontCreateGame(g)
	conn1 := createClient(t, buildGamePath(g))
	if conn1 == nil {
		return
	}
	h := []byte("{\"Status\":\"no_games_available\",\"OpponentPlay\":[\"\",\"\",\"\"],\"MyPlay\":[\"\",\"\",\"\"],\"Objective\":\"\",\"Clock\":0,\"Points\":0}")
	verifyReceive(t, g, conn1, h)
	conn1.Close()
	releaseBufferServer()
}

func TestGameFull(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g := "full_game"
	createGame(g, "Fullerene")
	conn1 := createClient(t, buildGamePath(g))
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, buildGamePath(g))
	if conn2 == nil {
		return
	}
	conn3 := createClient(t, buildGamePath(g))
	if conn3 == nil {
		return
	}
	h := []byte("{\"Status\":\"no_games_available\",\"OpponentPlay\":[\"\",\"\",\"\"],\"MyPlay\":[\"\",\"\",\"\"],\"Objective\":\"\",\"Clock\":0,\"Points\":0}")
	verifyReceive(t, g, conn3, h)
	conn1.Close()
	conn2.Close()
	conn3.Close()
	releaseBufferServer()
}

func verifyGameState(t *testing.T, msg []byte, g *game, pos position){
	actual_msg, err := json.Marshal(g.gameState(g.players[pos]))
	if err != nil {
		t.Errorf("Read: %s %v", g.gid, err)
	}
	if !bytes.Equal(msg, actual_msg) {
		t.Errorf("Echo: %s expected \n%q\n got \n%q", g.gid, msg, actual_msg)
	}
}
func TestGameBackspace(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g := newGame("backspace", "HO")
	g.objective = "HO"
	var v ws.Conn
	g.players[Fore].sock = &v
	g.players[Aft].sock = &v
	bkspmsg := []byte("{\"Name\":\"down\",\"KeyRune\":8}")
	hmsg := []byte("{\"Name\":\"down\",\"KeyRune\":72}")
	imsg := []byte("{\"Name\":\"down\",\"KeyRune\":73}")
	var h,i,bksp keypress 
	json.Unmarshal(hmsg, &h)
	json.Unmarshal(imsg, &i)
	json.Unmarshal(bkspmsg, &bksp)
	g.clock = 10
	g.integrate(g.players[Fore], h)
	
	verifyGameState(t,[]byte("{\"Status\":\"gaming\",\"OpponentPlay\":[\"\",\"\",\"HO\"],\"MyPlay\":[\"H\",\"\",\"O\"],\"Objective\":\"HO\",\"Clock\":10,\"Points\":0}"), g, Fore)
	g.integrate(g.players[Fore], i)
	verifyGameState(t,[]byte("{\"Status\":\"gaming\",\"OpponentPlay\":[\"\",\"\",\"HO\"],\"MyPlay\":[\"H\",\"I\",\"O\"],\"Objective\":\"HO\",\"Clock\":10,\"Points\":0}"), g, Fore)
	g.integrate(g.players[Fore], bksp)
	verifyGameState(t,[]byte("{\"Status\":\"gaming\",\"OpponentPlay\":[\"\",\"\",\"HO\"],\"MyPlay\":[\"H\",\"\",\"O\"],\"Objective\":\"HO\",\"Clock\":10,\"Points\":0}"), g, Fore)
	
}

/*
func TestGameReconnectConn(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	g:= "reconnect"
	createGame(g, "")
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
*/
