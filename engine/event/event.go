package event

import (
	"fmt"
)

// Event represents a game event that can be triggered either
// by the user or the engine itself.
type Event interface {
	Execute()
	fmt.Stringer
}

// A VisitorConnected is generated when a client connects to the lobby.
type VisitorConnected struct {
	Sock *ws.Conn
	// ...
}
// A VisitorDisconnected is generated when a client disconnects from the game.
type VisitorDisconnected struct {
	Sock *ws.Conn
	// ...
}
// A PlayerConnected is generated when a player identifies themself (logs in.)
type PlayerConnected struct {
	Player *Player
	// ...
}

// Execute satisfies the Event interface requirements.
func (p PlayerConnected) Execute() {
	// ...
}

// String satisfies the Stringer interface requirements.
func (p PlayerConnected) String() string {
	return "event:player_connected"
}

// A PlayerDisconnected is generated when a player logs out.
type PlayerDisconnected struct {
	Player *Player
}

// Execute satisfies the Event interface requirements.
func (p PlayerDisconnected) Execute() {
	// ...
}

// String satisfies the Stringer interface requirements.
func (p PlayerDisconnected) String() string {
	return "event:player_disconnected"
}
