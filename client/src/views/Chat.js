import { React, useState } from 'react'
import Form from './Form'

const skt_endpoint = 'ws://localhost:5678/ws';

function Chat({ videoID }) {

  var username = "";
  var userToken = {}
  var socket;


  const [messages, setMessages] = useState([])
  const [closed, toggleClosed] = useState(true);
  const [loaded, toggleLoaded] = useState(false);


  const addNewMessage = (msg) => {
    setMessages(messages.slice(Math.max(messages.length - 20, 0), messages.length).push(msg));
    console.log(messages);
  }

  const pushMessage = (msg) => {
    socket.send(JSON.stringify(msg));
  }

  const submitUsername = (inputUserName) => {
    username = inputUserName;

    userToken = { userName: inputUserName, videoID: videoID, userID: 0 }
    socket = new WebSocket(skt_endpoint);

    socket.onopen = (_) => {
      console.log(userToken);
      socket.send(JSON.stringify(userToken))
    }

    socket.onclose = (_) => {
      toggleClosed(true);
    }
    socket.onmessage = (ev) => {
      if (!loaded) {
        userToken.userID = JSON.parse(ev.data).userID;
        toggleLoaded(true);
        console.log(userToken)
      } else {
        addNewMessage(JSON.parse(ev.data))
      }
    }
  }

  if (!closed) {
    return (
      <div>
        Service unavailable. Restart the extension later...
      </div>
    )
  } else if (loaded) {
    return (
      <div>
        Chat should appear!
      </div>
    )
  } else {
    return (
      <Form videoID={videoID} submitUsername={submitUsername} isLoaded={loaded} />
    )
  }
}

export default Chat

