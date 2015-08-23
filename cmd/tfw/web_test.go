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
	http.Handle("/buffer", ws.Handler(bufferServer))
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

/*
func TestMultiEchoOneConn(t *testing.T) {

	once.Do(startServer)
	conn := createClient(t, "/echo")
	if conn == nil {
		return
	}

	msg := []byte("hello, world\n")
	if _, err := conn.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn, msg)
	conn.Close()
}
*/
func verifyReceive(t *testing.T, conn *ws.Conn, msg []byte) {
	var actual_msg = make([]byte, 512)

	n, err := conn.Read(actual_msg)
	if err != nil {
		t.Errorf("Read: %v", err)
	}
	actual_msg = actual_msg[0:n]
	if !bytes.Equal(msg, actual_msg) {
		t.Errorf("Echo: expected %q got %q", msg, actual_msg)
	}

}

/*
func TestMultiEchoTwoConn(t *testing.T) {

	once.Do(startServer)
	conn1 := createClient(t, "/echo")
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, "/echo")
	if conn2 == nil {
		return
	}

	msg := []byte("hello, world\n")
	if _, err := conn1.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn1, msg)
	verifyReceive(t, conn2, msg)

	conn1.Close()
	conn2.Close()
}
*/
/*
func TestMultiEchoCloseConn(t *testing.T) {

	once.Do(startServer)
	conn1 := createClient(t, "/echo")
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, "/echo")
	if conn2 == nil {
		return
	}

	msg := []byte("hello, world\n")
	if _, err := conn1.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn1, msg)
	verifyReceive(t, conn2, msg)

	conn1.Close()

	if _, err := conn2.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn2, msg)
	conn2.Close()
}
*/

func TestGameBackspace(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	conn1 := createClient(t, "/buffer")
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, "/buffer")
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

	verifyReceive(t, conn1, h)

	if _, err := conn1.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn1, hi)

	if _, err := conn1.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn1, h)
	conn1.Close()
	conn2.Close()
}

func TestGameReconnectConn(t *testing.T) {
	once.Do(startServer)
	initBufferServer()
	conn1 := createClient(t, "/buffer")
	if conn1 == nil {
		return
	}
	conn2 := createClient(t, "/buffer")
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

	verifyReceive(t, conn1, h1)
	verifyReceive(t, conn2, h2)

	if _, err := conn2.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn1, i1)
	verifyReceive(t, conn2, i2)

	conn1.Close()

	if _, err := conn2.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	que := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	verifyReceive(t, conn2, que)

	wait2 := []byte("{\"Status\":\"waiting_for_opponent\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	if _, err := conn2.Write(bkspmsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t, conn2, wait2)

	conn1 = createClient(t, "/buffer")
	if conn1 == nil {
		return
	}
	if _, err := conn1.Write(imsg); err != nil {
		t.Errorf("Write: %v", err)
	}
	wait1 := []byte("{\"Status\":\"gaming\",\"OpponentPlay\":\"\",\"MyPlay\":\"HI\"}")
	regame2 := []byte("{\"Status\":\"waiting_for_opponent\",\"OpponentPlay\":\"H\",\"MyPlay\":\"\"}")
	verifyReceive(t, conn1, wait1)
	verifyReceive(t, conn2, regame2)
	conn2.Close()
	conn1.Close()
}

/*
func TestBufferRace(t *testing.T) {
	once.Do(startServer)
	initBufferServer()

	var (
		conns  [10]*ws.Conn
		writes [10]string
		wg     sync.WaitGroup
	)

	msg := []byte(".")
	wg.Add(10)

	for i := range conns {
		go func(i int) {
			defer wg.Done()

			if conns[i] = createClient(t, "/buffer"); conns[i] == nil {
				return
			}
			defer conns[i].Close()

			if _, err := conns[i].Write(msg); err != nil {
				t.Errorf("Write: %v", err)
				return
			}

			actualMsg := make([]byte, 512)
			if n, err := conns[i].Read(actualMsg); err != nil {
				t.Errorf("Read: %v", err)
				return
			} else {
				writes[i] = string(actualMsg[0:n])
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("{" + strings.Join(writes[:], "") + "}")

	conn2 := createClient(t, "/buffer")
	if conn2 == nil {
		return
	}
	if _, err := conn2.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	message := []byte(".")
	for range conns {
		msg = append(msg, message[0])
	}
	verifyReceive(t, conn2, msg)
	conn2.Close()
}
*/
