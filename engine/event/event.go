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

// A PlayerConnected is generated when a client connects to
// play a game.
type PlayerConnected struct {
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

// A PlayerDisconnected is generated when a client disconnects
// from the server.
type PlayerDisconnected struct {
	// ...
}

// Execute satisfies the Event interface requirements.
func (p PlayerDisconnected) Execute() {
	// ...
}

// String satisfies the Stringer interface requirements.
func (p PlayerDisconnected) String() string {
	return "event:player_disconnected"
}
