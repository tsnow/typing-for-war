package visitor

import (
	"fmt"
	"log"

	"github.com/tsnow/typing-for-war/engine/event"
	le "github.com/tsnow/typing-for-war/logging/log"
	ws "code.google.com/p/go.net/websocket"	

)

// A Visitor represents an agent (human or otherwise) that can engage
// in gameplay with another agent.
type Visitor struct {
	sock *ws.Sock
}

// New returns an initialized Visitor structure.
func New(sock *ws.Sock) *Visitor {
	return &Visitor{
		sock: sock
	}
}
type visitorEventLog struct{
	Visitor *Visitor
	Err error
}

func (p *Visitor) Listen(){
	le.Event("visitor:connect", visitorEventLog{p,nil})
	var message string
	for {
		err := ws.Message.Receive(p.sock, &message)
		if err != nil {
			data := visitorEventLog{p,err}
			le.Event("visitor:receive_fail", data)
			
			p.sock.Close()
			le.Event("visitor:disconnected:from_fail", data)
			p.engine.Disconnect(p.sock)
			break
		}
		p.received(message)
	}
}

func (p *Visitor) received(message string){
	p.engine.
}

// Update is called by the engine when an event occurs.
func (p *Visitor) Update(e event.Event) {
	log.Printf("%s received event:%s\n", p, e)
	
}

// String satisfies the Stringer interface requirements.
func (p *Visitor) String() string {
	return fmt.Printf("visitor: %s", p.sock.RemoteAddr())
}
