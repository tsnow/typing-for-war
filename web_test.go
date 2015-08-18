package main

import (
	ws "code.google.com/p/go.net/websocket"
	"bytes"
	"fmt"

	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"strings"
	"sync"
	"testing"
	"time"
)

var serverAddr string
var once sync.Once

func startServer() {
	initMultiEcho()
	http.Handle("/echo", ws.Handler(multiEchoServer))
	http.Handle("/buffer", ws.Handler(bufferServer))
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
}

func newConfig(t *testing.T, path string) *ws.Config {
	config, _ := ws.NewConfig(fmt.Sprintf("ws://%s%s", serverAddr, path), "http://localhost")
	return config
}

func createClient(t *testing.T, resource string) *ws.Conn{

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

func TestMultiEchoOneConn(t *testing.T){

	once.Do(startServer)
	conn := createClient(t, "/echo")
	if conn == nil {
		return;
	}
	
	msg := []byte("hello, world\n")
	if _, err := conn.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn,msg)
	conn.Close()
}

func verifyReceive(t *testing.T, conn *ws.Conn, msg []byte){
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

func TestMultiEchoTwoConn(t *testing.T){

	once.Do(startServer)
	conn1 := createClient(t,"/echo")
	if conn1 == nil {
		return;
	}
	conn2 := createClient(t,"/echo")
	if conn2 == nil {
		return;
	}
	
	msg := []byte("hello, world\n")
	if _, err := conn1.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn1,msg)
	verifyReceive(t,conn2,msg)
	
	conn1.Close()
	conn2.Close()
}
func TestMultiEchoCloseConn(t *testing.T){

	once.Do(startServer)
	conn1 := createClient(t, "/echo")
	if conn1 == nil {
		return;
	}
	conn2 := createClient(t, "/echo")
	if conn2 == nil {
		return;
	}
	
	msg := []byte("hello, world\n")
	if _, err := conn1.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn1,msg)
	verifyReceive(t,conn2,msg)
	
	conn1.Close()

	if _, err := conn2.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn2,msg)
	conn2.Close()
}


func TestBufferCloseConn(t *testing.T){
	once.Do(startServer)
	initBufferServer()
	conn1 := createClient(t,"/buffer")
	if conn1 == nil {
		return;
	}
	conn2 := createClient(t,"/buffer")
	if conn2 == nil {
		return;
	}
	
	conn1msg := []byte("hello, ")
	conn2msg := []byte("world \n")
	combinedMsg := []byte("hello, world \n")
	if _, err := conn1.Write(conn1msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn1,conn1msg)
	verifyReceive(t,conn2,conn1msg)
	if _, err := conn2.Write(conn2msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	verifyReceive(t,conn1,combinedMsg)
	verifyReceive(t,conn2,combinedMsg)
	
	conn1.Close()

	if _, err := conn2.Write([]byte("goodnight moon")); err != nil {
		t.Errorf("Write: %v", err)
	}
	combinedMsg = []byte("hello, world \ngoodnight moon")
	verifyReceive(t,conn2,combinedMsg)
	conn2.Close()
}

func TestBufferRace(t *testing.T){
	once.Do(startServer)
	initBufferServer()
	conns := [10]*ws.Conn{}
	writes := [10]string{}
	msg := []byte(".")
	done:= make(chan bool)
	for i := range conns {
		go func(i int){
			defer func(){ done <- true }()
			j := i;
			conns[j] = createClient(t,"/buffer")
			if conns[j] == nil {
				return;
			}
			defer conns[j].Close()
			if _, err := conns[j].Write(msg); err != nil {
				t.Errorf("Write: %v", err)
				return;
			}
			var actual_msg = make([]byte, 512)
// 			conns[j].SetReadDeadline(time.Now().Add(time.Second))
			n, err := conns[j].Read(actual_msg)
			if err != nil {
				t.Errorf("Read: %v", err)
				return;
			}
			actual_msg = actual_msg[0:n]
			writes[j] = string(actual_msg)
		}(i)
	}
	go func(){
		i := 0
		for {
			select {
			case <- done:
				i++
				if i >= len(conns) {
					break;
				}
			case <-time.After(1*time.Second):
				t.Errorf("ran out of time")
				close(done)
				break;
			}
		}
	}()
	<- done
	var b []string
	b = writes[:]
	fmt.Printf("{"+strings.Join(b, "_")+"}")
	conn2 := createClient(t,"/buffer")
	if conn2 == nil {
		return;
	}
	if _, err := conn2.Write(msg); err != nil {
		t.Errorf("Write: %v", err)
	}
	message := []byte("?")
	for range conns {
		msg = append(msg, message[0])
	}
	verifyReceive(t,conn2,msg)
	conn2.Close()
}

