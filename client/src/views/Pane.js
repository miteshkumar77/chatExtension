import React from 'react'
import { Card, CardBody, CardTitle, CardSubtitle, Badge } from "shards-react";
import "bootstrap/dist/css/bootstrap.min.css";
import "shards-ui/dist/css/shards.min.css"

const colors = ["secondary", "success", "info", "warning", "danger", "light", "dark"];
// own ? "text-align:right" : "text-align:left"
function Pane({ own, msg, username, color }) {
  return (
    <Card>
      <CardBody>
        <Badge color={colors[color]}>{username}</Badge>
        {"  "}{msg}
      </CardBody>
    </Card>
  );
}

export default Pane
