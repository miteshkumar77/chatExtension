let socket = new WebSocket("ws://localhost:5678/ws")
socket.onmessage = (e) => {
    console.log(e.data);
}
socket.onopen = (e) => {
    socket.send(JSON.stringify({ userName: "usrnm2", videoID: "ax45fpj9", userID: 0 }))
}
socket.send(JSON.stringify({ msgText: "Hi usrnm!", userName: "usrnm2", userID: 91616, timeStamp: "apples", videoID: "ax45fpj9" }))