package main

import (
	"fmt"
	"net/http"
	"os"

	ws "code.google.com/p/go.net/websocket"
	"log"

	"html/template"
)

    // OO UP IN THIS BITCH
    type GameList struct {
    	active []*Game
    	waiting *Game
    }

    func (this *GameList) Connect(conn *ws.Conn) *Game {
    	// if waiting is empty, create a new waiting game
    	// if waiting has any, FIFO complete game
    	if this.waiting == nil {
    		game := New()
    		game.Connect(conn)
    		this.waiting = game
    	} else {
    		this.waiting.Connect(conn)
    		this.active = append(this.active,this.waiting)
    		this.waiting.Broadcast("A game done did begunnned.")
    		return this.waiting
    		//this.waiting = nil
    	}
    	return this.waiting
    }

	type Game struct {
		first *ws.Conn
		last *ws.Conn
		disconnect chan bool
	}

	func New() (game *Game) {
		game = new(Game)
		game.disconnect = make(chan bool)
		go func(){
			for {
				select {
					case disconnect := <- game.disconnect:
						if disconnect {
							game.first.Close()
							game.last.Close()
							log.Print("- game ", game.first.RemoteAddr(), "<->",game.last.RemoteAddr(), " disconnected")
						}
				}
			}
		}()
		return
	}
	func (this *Game) Clone() (game *Game){
		game.first = this.first
		game.last = this.last
		game.disconnect = this.disconnect
		return
	}

	func (this *Game) Connect(conn *ws.Conn) {
		if this.first != nil {
			log.Print("- game last ", conn.RemoteAddr(), " connected")
			this.last = conn
		} else {
			log.Print("- game first ", conn.RemoteAddr(), " connected")
			this.first = conn
		}
	}

	func (this *Game) Disconnect() {
		this.disconnect <- true
	}

	func (this *Game) Broadcast (message string) {
		if ws.Message.Send(this.first,message) != nil {
			this.Disconnect()
		}
		if ws.Message.Send(this.last, message) != nil {
			this.Disconnect()
		}

	}

	var games *GameList
	// OO DONE 
/*
	type game struct {
		conns []*ws.Conn
		disconnect chan bool
	}
	var	current_game = &game{disconnect: make(chan bool)}
	func disconnectOnError(current_game *game) {
	go func(){
		for {
			select {
				case disconnect := <- current_game.disconnect:
					if disconnect {
						for _,conn := range current_game.conns {
							conn.Close()
							log.Print("- game ", conn.RemoteAddr(), " disconnected")
						}
						current_game.conns = nil
					}
			}
		}
	}()
	}
*/

func main() {
/* Blood and destruction shall be so in use 
And dreadful objects so familiar 
That mothers shall but smile when they behold 
Their infants quarter'd with the hands of war; 
All pity choked with custom of fell deeds: 
And Caesar's spirit, ranging for revenge, 
With Ate by his side come hot from hell, 
Shall in these confines with a monarch's voice 
Cry 'Havoc,' and let slip the dogs of war; 
That this foul deed shall smell above the earth 
With carrion men, groaning for burial. */

	http.HandleFunc("/app/index", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res,req,"/app/index.html") // /app/index.html for heroku
	})
	http.Handle("/", http.FileServer(http.Dir(os.Getenv("PWD")))) 
	http.Handle("/socket/echo", ws.Handler(func(sock *ws.Conn){
		log.Print("- ", sock.RemoteAddr(), " connected")
		var message string
		if ws.Message.Receive(sock, &message) != nil {
			log.Print("- ", sock.RemoteAddr(), " couldn't receive.")
		}

		if ws.Message.Send(sock,message) != nil {
			log.Print("- ", sock.RemoteAddr(), " couldn't send.")
		}
		sock.Close()
		log.Print("- ", sock.RemoteAddr(), " disconnected")
	}))

	// disconnectOnError(current_game)

	http.Handle("/socket/new_game", ws.Handler(func(sock *ws.Conn){
		game :=	 games.Connect(sock)

		var message string
		if ws.Message.Receive(sock, &message) != nil {
			log.Print("- game ", sock.RemoteAddr(), " couldn't receive.")
			game.disconnect <- true
		}
		game.Broadcast(message)
	}))


	template.New("things")
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

