import React, { useState, createRef, useEffect } from 'react'
import './Chat.css';
import Pane from './Pane';
import {
  InputGroup,
  InputGroupText,
  InputGroupAddon,
  FormInput,
  Button,
  Form,
} from "shards-react";
function Chat({ userID, userName, videoID, sendMessageFunc, messageList }) {

  const [messageText, setMessageText] = useState('');
  const bottom = createRef();

  const scrollToBottom = () => {
    bottom.current.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(scrollToBottom, [messageList]);

  return (
    <div>
      <main className="messagesWindow">
        {
          messageList.map((message, index) => {
            return (<Pane
              messageText={message.msgText}
              username={message.userName}
              own={userID === message.userID}
              timestamp={message.timeStamp}
              key={index}
            />);
          })
        }
        <div ref={bottom} />
      </main>
      <Form onSubmit={(e) => {
        e.preventDefault();
        if (messageText.length === 0) return;
        sendMessageFunc(messageText);
        setMessageText("");
      }}>
        <InputGroup >
          <InputGroupAddon type="prepend">
            <InputGroupText>
              {videoID}/
            </InputGroupText>
          </InputGroupAddon>
          <FormInput placeholder="type a message" value={messageText} onChange={(e) => { setMessageText(e.target.value) }} />
          <InputGroupAddon type="append">
            <Button theme="secondary">Send</Button>
          </InputGroupAddon>
        </InputGroup>
      </Form>
    </div>
  )
}

export default Chat
