package player

import (
	"github.com/tsnow/typing-for-war/engine/event"
	"github.com/tsnow/typing-for-war/logging/log"
)

// A Player represents an agent (human or otherwise) that can engage
// in gameplay with another agent.
type Player struct {
	Name string
}

// New returns an initialized Player structure.
func New(name string) *Player {
	return &Player{
		Name: name,
	}
}

// Update is called by the engine when an event occurs.
func (p *Player) Update(e event.Event) {
	log.Printf("%s received event:%s\n", p, e)
}

// String satisfies the Stringer interface requirements.
func (p *Player) String() string {
	return "player:" + p.Name
}
