import { React, useState } from 'react'
import { Card, CardBody, CardTitle, CardSubtitle, FormInput, Button, Alert } from "shards-react";
import "bootstrap/dist/css/bootstrap.min.css";
import "shards-ui/dist/css/shards.min.css"


function isAlNum(str) {
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


function Form({ videoID, submitUsername, isLoaded }) {

  const [userName, changeUserName] = useState("")

  return (
    <Card>
      <CardBody>
        <CardTitle>Video ID: {videoID}</CardTitle>
        <CardSubtitle>Enter nickname to join chat...</CardSubtitle>
        <FormInput placeholder="Enter nickname" onChange={(e) => { changeUserName(e.target.value) }} />
        {
          (() => {
            if (!isAlNum(userName)) {
              return (

                <div>
                  <Button theme="light" disabled={true}>Submit</Button>
                  <Alert theme="primary">
                    User name must consist of alphanumeric characters (0-9), (A-Z), (a-z)...
                  </Alert>
                </div>
              )

            }

            if (!(userName.length > 2)) {
              return (<div>
                <Button theme="light" disabled={true}>Submit</Button>
                <Alert theme="primary">
                  User name must be atleast 3 characters...
                </Alert>
              </div>)
            }

            if (!(userName.length < 16)) {
              return (<div>
                <Button theme="light" disabled={true}>Submit</Button>
                <Alert theme="primary">
                  User name cannot be more than 16 characters...
                </Alert>
              </div>)
            }

            return <Button theme="success" onClick={() => { submitUsername(userName) }}>Continue</Button>
          })()
        }
      </CardBody>
    </Card>
  )
}

export default Form
