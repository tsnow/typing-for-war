$(document).ready(function(){

    var player = function(name,host,container){
	this.container = container;
	this.name = name;
	this.host = host;
	this.tickerData="";
	try{
	    var websocket = new WebSocket(host);
	    this.websocket = websocket;
	    this.registerSocket();
	    this.registerContainer();
	} catch(exception){
	    this.message('<p>LoadError '+exception);
	}
    };
    player.prototype = {
	"ticker": function(){
	    var that=this;
	    that.container.find(".ticker").text(that.tickerData);
	},
	"registerSocket": function(){
	    var that=this;
	    try{
		
		that.message('<p class="event"> ' + that.name +' Socket Status: '+that.websocket.readyState);
		
		that.websocket.onopen = function(){
		    that.message('<p class="event"> ' + that.name + ' Socket Status: '+that.websocket.readyState+'(open)');
		}
		
		that.websocket.onmessage = function(msg){
		    that.message('<p class="message"> ' + that.name + ' Received: '+msg.data);
		}
		
		that.websocket.onclose = function(){
		    that.message('<p class="event"> ' + that.name + ' Socket Status: '+that.websocket.readyState+'(Closed)');
		}
		that.websocket.onerror = function(e){
		    that.message('<p class="event"> ' + that.name + ' Socket Status: '+that.websocket.readyState+'(Error):' + e);
		}
	    } catch(exception){
		that.message('<p> ' + that.name + ' Error'+exception);
	    }
	    
	},
	"registerContainer": function(){
	    var that=this;
	    var handler=function(name){
		var n = name;
		return function(key){
		    key.preventDefault();
		    // var text = that.container.find('#text').val();
		    if(n === "down"){
			var keypressed = key.keyCode <= 64 ? "^"+String.fromCharCode(key.keyCode + 64) : String.fromCharCode(key.keyCode)
			that.tickerData=that.tickerData+keypressed;
			that.ticker();
		    }

		    that.send(JSON.stringify([
			n,
			key.keyCode,
			String.fromCharCode(key.keyCode),
			key.keyCode <= 64 ? "^"+String.fromCharCode(key.keyCode + 64) : "",
			key.charCode
		    ]));
		};
	    };
	    that.container.find('#text').keydown(handler('down')).keypress(handler('press')).keyup(handler('up'));
	    that.container.find('#disconnect').click(function(){
		that.websocket.close();
	    });
	},
	"send": function(text){
	    if(text==""){
		this.message('<p class="warning">Please enter a message');
		return ;
	    }
	    try{
		this.websocket.send(text);
		this.message('<p class="event">Sent: '+text)
	    } catch(exception){
		this.message('<p class="warning"> Error:' + exception);
	    }

	},
	"message": function(msg){
	    this.container.find('#chatLog').append(msg+'</p>');
	}
    }

    // var host = "ws://typing-for-war.herokuapp.com/socket/new_game";
    // var host = "ws://phaedo.local:5002/socket/new_game";
    var host = "ws://localhost:5002/socket/buffer"; 
    var player1 = new player('Player 1', host, $('.player1'));
    var player2 = new player('Player 2', host, $('.player2'));
});
