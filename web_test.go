package main

import (
	ws "code.google.com/p/go.net/websocket"
	"bytes"
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
	initMultiEcho()
	http.Handle("/echo", ws.Handler(multiEchoServer))
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
}

func newConfig(t *testing.T, path string) *ws.Config {
	config, _ := ws.NewConfig(fmt.Sprintf("ws://%s%s", serverAddr, path), "http://localhost")
	return config
}
func createClient(t *testing.T) *ws.Conn{

	// websocket.Dial()
	client, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	conn, err := ws.NewClient(newConfig(t, "/echo"), client)
	if err != nil {
		t.Errorf("WebSocket handshake error: %v", err)
		return nil
	}
	return conn
}
func TestMultiEchoOneConn(t *testing.T){

	once.Do(startServer)
	conn := createClient(t)
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
	conn1 := createClient(t)
	if conn1 == nil {
		return;
	}
	conn2 := createClient(t)
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
	conn1 := createClient(t)
	if conn1 == nil {
		return;
	}
	conn2 := createClient(t)
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

