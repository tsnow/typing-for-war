
    var player = function(name,host,container){
        this.container = container;
        this.name = name;
        this.host = host;
        this.tickerData="";
        try{
            this.registerContainer();
        } catch(exception){
            this.message('<p>LoadError '+exception);
        }
    };
    player.prototype = {
        "ticker": function(){
            var that=this;
            that.container.find(".ticker").html(that.tickerData);
        },
        "registerSocket": function(){
            var that=this;
            var websocket = new WebSocket(that.host);
            this.websocket = websocket;
            try{
                
                that.message('<p class="event"> ' + that.name +' Socket Status: '+that.websocket.readyState);
                
                that.websocket.onopen = function(){
                    that.message('<p class="event"> ' + that.name + ' Socket Status: '+that.websocket.readyState+'(open)');
                }
                
                that.websocket.onmessage = function(msg){
                    that.onGameStateUpdate(msg.data);
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
        "playDisplay": function(playState){
            if(playState.join("") === ""){
                return "Loading...";
            }
            return '[<span class="playCorrect">'+playState[0]+'</span>'+
                '<span class="playWrong">'+playState[1]+'</span>'+
                '<span class="playLeft">'+playState[2]+'</span>]';
        },
        "onGameStateUpdate": function(msg){
            var that = this;
            try {
                var gameState = JSON.parse(msg);
            } catch(e) {
                console.log(msg);
                console.log(e);
                return;
            }
            that.tickerData = '';
            that.tickerData = that.tickerData + '<div class="event">Opponent: '+that.playDisplay(gameState.OpponentPlay)+'</div>';
            that.tickerData = that.tickerData + '<hr />';
            that.tickerData = that.tickerData + '<div class="warning">Game State: '+gameState.Status+'</div>';
            that.tickerData = that.tickerData + '<div class="warning">Countdown: '+gameState.Clock+'</div>';
            that.tickerData = that.tickerData + '<div class="warning"> Points: '+gameState.Points+'</div>';
            
            that.tickerData = that.tickerData + '<hr />';
            //that.tickerData = that.tickerData + '<div class="warning">Objective: '+gameState.Objective+'</p>';

            that.tickerData = that.tickerData + '<div class="message">CHALLENGE: '+that.playDisplay(gameState.MyPlay)+'</div>';
            that.ticker();
        },
        "connect": function(){
            this.registerSocket();
        },
        "disconnect": function(){
            this.websocket.close();
            this.websocket = undefined;
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

                    that.send(JSON.stringify({
                        Name: n,
                        KeyRune: key.keyCode,
                        CharRune: key.charCode
                    }));
                };
            };
            that.container.keydown(handler('down')).keypress(handler('press')).keyup(handler('up'));
            that.container.find('#disconnect').click(function(){
                if(that.websocket == undefined){
                    that.connect();
                    that.container.find('#disconnect').text("Disconnect");
                    that.container.find('#text').focus();
                }else{
                    that.disconnect();
                    that.container.find('#disconnect').text("Connect");
                }
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
        },
        "toggleDebugLog": function(){
            var that = this;
            if(that.debugLogShown == undefined || that.debugLogShown){
                that.container.find("#chatLog").hide();
                //that.container.find(".ticker").hide();
                that.debugLogShown = false;
            } else {
                that.container.find("#chatLog").show();
                //that.container.find(".ticker").show();
                that.debugLogShown = true;
            }
        }
    }

$(document).ready(function(){

    var wshost = window.location.host;
    var ws =  'ws://' + wshost + '/game/sparklemotion';
    // var host = "ws://typing-for-war.herokuapp.com/game/sparklemotion";
    // var host = "ws://phaedo.local:5002/game/sparklemotion";
    // var host = "ws://localhost:5002/game/sparklemotion"; 
    var player1 = new player('Player 1', ws, $('.player1'));
    var player2 = new player('Player 2', ws, $('.player2'));
    player1.toggleDebugLog();
    player2.toggleDebugLog();
    var player2hidden = false;
    var player2toggle = function(){
        if(player2hidden){
            player2.container.show();
        }else{
            player2.container.hide();
        }
        player2hidden = !player2hidden;
    };
    player2toggle();
    $('.toggleDebugLog').click(function(){
        player1.toggleDebugLog();
        player2.toggleDebugLog();
        player2toggle();
    });

});
