
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
	    $.each(gameState.Actions, function(ind,it){
		if(!library[it]){
		    return;
		}
		sfx[it]();
	    });
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
var library = {
    "gamestart":{
	"Frequency":{"Start":212.33506640419364,"Min":758.8429812574759,"Max":574.8486667172983,"Slide":-0.9446392799727619,"DeltaSlide":-0.7098942114971578,"RepeatSpeed":0.7800259343348444,"ChangeAmount":5.780829096212983,"ChangeSpeed":0.7770580844953656},"Vibrato":{"Depth":0.48443294526077807,"DepthSlide":-0.3592237173579633,"Frequency":22.47475597843761,"FrequencySlide":-0.8084569782949984},"Generator":{"Func":"noise","A":0.3764359597116709,"B":0.3190609645098448,"ASlide":0.00663434574380517,"BSlide":0.5758746429346502},"Guitar":{"A":0.5767553918994963,"B":0.05008110054768622,"C":0.0892909886315465},"Phaser":{"Offset":0.4939028648659587,"Sweep":0.35900392988696694},"Volume":{"Master":0.4,"Attack":0.9609616734087467,"Sustain":0.173606735188514,"Punch":0.3720094033051282,"Decay":1.232456877361983}},
    "gamelose":{
	"Frequency":{"Start":813.5479682683945,"Min":110,"Slide":-0.8397897131741047},"Generator":{"Func":"saw","A":0.7496007442008704,"ASlide":0.38207407603040333},"Phaser":{"Offset":0.008988518221303822,"Sweep":0.06145181683823467},"Volume":{"Sustain":0.11809764727950096,"Decay":0.2351402602158487}},
    "gamewin":{
	"Frequency":{"Start":1410.544993430376,"Slide":-0.3857578541617841,"RepeatSpeed":0.6960926694562659,"ChangeSpeed":0.8800750505179167,"ChangeAmount":0.17242067120969296},"Generator":{"Func":"noise"},"Phaser":{"Offset":0.12330289890524004,"Sweep":-0.2787581878714263},"Volume":{"Sustain":0.37970589983742686,"Decay":0.17686830135062337,"Punch":0.7969077744055539}},
    "memberlogin":{
	"Frequency":{"Start":656.9257708825171,"Min":221.7122994409874,"Max":1570.9409914864227,"Slide":-0.00766729936003685,"DeltaSlide":0.8205740465782583,"RepeatSpeed":1.3638154298532754,"ChangeAmount":6.730239121243358,"ChangeSpeed":0.20631370320916176},"Vibrato":{"Depth":0.7825833428651094,"DepthSlide":0.858677841257304,"Frequency":7.364898596368731,"FrequencySlide":-0.509336469694972},"Generator":{"Func":"saw","A":0.9032189801800996,"B":0.03754007490351796,"ASlide":-0.9324617250822484,"BSlide":0.6839395263232291},"Guitar":{"A":0.6332177349831909,"B":0.9940286141354591,"C":0.3287406745366752},"Phaser":{"Offset":-0.5544541538693011,"Sweep":0.15606358228251338},"Volume":{"Master":0.4,"Attack":0.9510696434881538,"Sustain":1.8074330082163215,"Punch":0.014344010967761278,"Decay":0.719868098385632}},
    "memberlogoff":{
	"Frequency":{"Start":1180},"Vibrato":{"Frequency":0.01,"Depth":0},"Generator":{"A":0,"B":0},"Filter":{"LP":0.88,"LPSlide":-0.02,"LPResonance":0,"HP":0},"Phaser":{"Offset":0.04,"Sweep":0.04},"Volume":{"Attack":0,"Punch":0.48,"Sustain":0.1,"Master":0.21}},
    "playerattack":{
	"Frequency":{"Start":239.96915185358375},"Generator":{"Func":"saw","A":0.13543841773644089},"Filter":{"HP":0.2},"Volume":{"Sustain":0.1341964266030118,"Decay":0.0714280589018017}},
    "gamecomplete":{
	"Frequency":{"Start":627.2097676480189,"Slide":0.21597476182505487},"Generator":{"Func":"saw"},"Volume":{"Sustain":0.33077970324084166,"Decay":0.4262595309875906}},
    "charplayed":{
	"Frequency":{"Start":1526,"Min":978,"Slide":0.77,"DeltaSlide":0.51,"RepeatSpeed":2.88,"ChangeAmount":9,"ChangeSpeed":0.84},"Generator":{"Func":"unoise","A":0.48,"B":0,"ASlide":-1,"BSlide":-0.93},"Filter":{"HP":0.15,"LPResonance":0,"LPSlide":-0.08},"Volume":{"Sustain":0,"Decay":0.1,"Attack":0,"Punch":0.04,"Master":0.71},"Vibrato":{"Depth":0,"Frequency":0.01,"DepthSlide":-0.03},"Phaser":{"Offset":-1,"Sweep":-1}},
    "countdown":{
	"Generator":{"B":1,"A":0},"Volume":{"Decay":0,"Punch":0,"Sustain":0.04}},
    "timepassing":{
	"Generator":{"B":1,"A":0,"Func":"sine","ASlide":-0.81},"Volume":{"Decay":0,"Punch":0,"Sustain":0.04}},
    "timerunningout":{
	"Generator":{"B":1,"A":0.83,"Func":"string","ASlide":-0.81},"Volume":{"Decay":0,"Punch":2.82,"Sustain":0.28,"Attack":0.02}}
};
var sfx = jsfx.Sounds(library);

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
