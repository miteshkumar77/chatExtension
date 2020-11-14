import { React, useState, useRef, useEffect } from 'react'
import { Card, CardBody, CardTitle, CardSubtitle, FormInput, Button, Alert } from "shards-react";

import Form from './Form'
import Pane from './Pane'
const skt_endpoint = 'ws://localhost:5678/ws';
var socket;
var userToken = {}
var isLoaded = false;

function Chat({ videoID }) {

  var username = "";

  // { msgText: "Hi usrnm!", userName: "usrnm2", userID: 91616, timeStamp: "apples", videoID: "ax45fpj9" }
  const [messages, setMessages] = useState([])
  const [closed, toggleClosed] = useState(true);
  const [loaded, toggleLoaded] = useState(false);
  const [currentMessage, setCurrentMessage] = useState();
  const fieldRef = useRef(null);

  // useEffect(() => {
  //   if (fieldRef.current != null) {
  //     window.scrollTo(0, fieldRef.current.offsetTop)
  //   }
  // }, [messages])

  const addNewMessage = (msg) => {
    console.log("adding message")
    // console.log(`first: ${Math.max(messages.length - 20, 0)}  second: ${messages.length}...`)
    console.log([...messages.slice(Math.max(messages.length - 20, 0), messages.length), msg])
    setMessages([...messages.slice(Math.max(messages.length - 20, 0), messages.length), msg]);
    console.log(messages);
  }

  const pushMessage = (msg) => {
    console.log(socket);
    console.log(msg);
    socket.send(JSON.stringify(msg));
  }

  const submitUsername = (inputUserName) => {
    username = inputUserName;

    userToken = { userName: inputUserName, videoID: videoID, userID: 0 }
    socket = new WebSocket(skt_endpoint);

    socket.onopen = (_) => {
      console.log(userToken);
      console.log(socket)
      socket.send(JSON.stringify(userToken))
    }

    socket.onclose = (_) => {
      toggleClosed(true);
    }
    socket.onmessage = (ev) => {
      console.log("DATA")
      console.log(ev.data)
      console.log(loaded)
      if (!isLoaded) {
        userToken.userID = JSON.parse(ev.data).userID;
        (() => {
          isLoaded = true;
          toggleLoaded(true);
        })()
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
        {/* styles="overflow:scroll" */}
        <div >
          {

            messages.map((msg) => {
              return (
                <Pane
                  own={msg.userID == userToken.userID}
                  msg={msg.msgText}
                  username={msg.userName}
                  color={0}
                />
              );
            })
          }
          <div className="field" ref={fieldRef} />
        </div>
        <FormInput placeholder="Enter nickname" onChange={(e) => { setCurrentMessage(e.target.value) }} />
        <Button theme="success" onClick={() => {

          const msg_to_send = {
            msgText: currentMessage,
            userName: username,
            userID: userToken.userID,
            timeStamp: "hello",
            videoID: userToken.videoID
          }
            (() => {
              pushMessage(msg_to_send);
              setCurrentMessage("");
            })();
        }}>Continue</Button>
      </div>
    )
  } else {
    return (
      <Form videoID={videoID} submitUsername={submitUsername} isLoaded={loaded} />
    )
  }
}

export default Chat

