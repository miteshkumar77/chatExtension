import React from 'react'
import {
  InputGroup,
  InputGroupText,
  InputGroupAddon,
  FormInput,
  Button,
  Form,
  Alert
} from "shards-react";
import Chat from './Chat';
import "bootstrap/dist/css/bootstrap.min.css";
import "shards-ui/dist/css/shards.min.css"
const url = 'ws://localhost:8080/ws'
const isAlNum = (str) => {
  var code, i, len;

  for (i = 0, len = str.length; i < len; i++) {
    code = str.charCodeAt(i);
    if (!(code > 47 && code < 58) && // numeric (0-9)
      !(code > 64 && code < 91) && // upper alpha (A-Z)
      !(code > 96 && code < 123)) { // lower alpha (a-z)
      return false;
    }
  }
  return true;
}


const isValidUsername = (userName) => {
  if (!isAlNum(userName)) {
    return "User name must consist of alphanumeric characters (0-9), (A-Z), (a-z)...";
  } else if (!(userName.length > 2)) {
    return "User name must be atleast 3 characters...";
  } else if (!(userName.length < 16)) {
    return "User name cannot be more than 16 characters...";
  } else {
    return "Connecting to room...";
  }
}

let globalSock;

export default class Main extends React.Component {

  constructor(props) {
    super(props);
    this.state = {
      messages: [],
      userNameText: '',
      signInMessage: "Enter a username to join...",
      signedIn: false,
      credentials: null,
    }
  }

  onOpenHandler = (_) => {
    globalSock.send(JSON.stringify({
      userName: this.state.userNameText,
      videoID: this.props.videoID,
      userID: 0
    }));
  }

  onMessageHandler = (e) => {
    if (this.state.signedIn === false) {
      console.log("Not signed in: ");
      console.log(e.data);
      this.setState({
        credentials: JSON.parse(e.data),
        signedIn: true
      });
    } else {
      let start = Math.max(0, this.state.messages.length - 10);
      let incomingMessage = JSON.parse(e.data);
      console.log(e.timestamp);
      // incomingMessage.timeStamp = String.toString(e.timeStamp);
      console.log("incoming message: ");
      console.log(incomingMessage);
      this.setState({
        messages: [
          ...this.state.messages.slice(start, this.state.messages.length),
          incomingMessage
        ]
      });
    }
  }

  onCloseHandler = (e) => {
    console.log("Connection closed...");
    this.setState({
      signInMessage: "An error ocurred, try signing in again.",
      messages: [],
      signedIn: false,
      credentials: null
    })
  }

  formSubmit = (e) => {
    e.preventDefault();
    this.setState({
      signInMessage: isValidUsername(this.state.userNameText)
    }, () => {
      if (this.state.signInMessage === "Connecting to room...") {
        globalSock = new WebSocket(url);
        globalSock.onmessage = this.onMessageHandler;
        globalSock.onopen = this.onOpenHandler;
        globalSock.onclose = this.onCloseHandler;
      }
    })
  }

  sendMessage = (messageText) => {
    if (this.state.credentials === null) {
      console.log("Credentials haven't been set yet!");
      return;
    }

    globalSock.send(JSON.stringify({
      msgText: messageText,
      ...this.state.credentials,
      timeStamp: ""
    }))
  }

  render() {
    const { signInMessage, credentials, messages } = this.state;
    const { videoID } = this.props
    if (this.state.signedIn === false) {
      return (
        <div>
          {(signInMessage !== "Connecting to room...") && (
            <Form onSubmit={this.formSubmit}>
              <InputGroup >
                <InputGroupAddon type="prepend">
                  <InputGroupText>
                    {videoID}/
            </InputGroupText>
                </InputGroupAddon>
                <FormInput
                  placeholder="username"
                  onChange={(e) => { this.setState({ userNameText: e.target.value }) }}
                />
                <InputGroupAddon type="append">
                  <Button theme="secondary">Enter</Button>
                </InputGroupAddon>
              </InputGroup>
            </Form>)}
          <Alert>{signInMessage}</Alert>
        </div>
      )
    } else {
      return (
        <Chat
          userID={credentials.userID}
          userName={credentials.userName}
          videoID={credentials.videoID}
          sendMessageFunc={this.sendMessage}
          messageList={messages}
        />
      )
    }
  }
}
