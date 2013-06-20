package main

import (
	"fmt"
	"net/http"
	"os"

	ws "code.google.com/p/go.net/websocket"
	"log"

	"html/template"
)

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

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res,req,"/app/index.html")
	})
	http.Handle("/files", http.FileServer(http.Dir(os.Getenv("PWD"))))
	http.Handle("/socket/server", ws.Handler(func(sock *ws.Conn){
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
	template.New("things")
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

