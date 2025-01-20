 
# websocket
 

create a client for test

python3 -m http.server 8000




const socket = new WebSocket("ws://localhost:8080/ws");

socket.onopen = function() {
    console.log("WebSocket bağlantısı kuruldu!");
};

socket.onmessage = function(event) {
    console.log("Sunucudan gelen mesaj:", event.data);
};

socket.onclose = function() {
    console.log("WebSocket bağlantısı kapatıldı.");
};