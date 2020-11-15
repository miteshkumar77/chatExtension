import React from 'react'
import { Card, CardBody, Badge, Alert } from "shards-react";
import "bootstrap/dist/css/bootstrap.min.css";
import "shards-ui/dist/css/shards.min.css"

const themes = ["secondary", "success", "info", "warning", "danger", "light", "dark"];
// own ? "text-align:right" : "text-align:left"
function Pane({ own, messageText, username, timestamp }) {
  return (
    // <Card>
    //   <CardBody>
    <Alert theme={themes[own ? 0 : 1]}>
      <Badge theme="light" style={{ marginRight: '10px' }}>{username}</Badge>
      {"       "}{messageText}
    </Alert>
    //   </CardBody>
    // </Card>
  );
}

export default Pane
