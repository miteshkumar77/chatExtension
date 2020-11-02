// import { Widget } from "react-chat-widget";

// import "react-chat-widget/lib/styles.css";
/* <Widget
  handleNewUserMessage={handleNewUserMessage}
  title={`video: ${videoID}`}
  subtitle=""
/> */
/*global chrome*/

function Chat({ videoID }) {
  const handleNewUserMessage = (newMessage) => {
    console.log(`New message incoming! ${newMessage}`);
  };
  return (
    <div>

    </div>
  );
}

export default Chat;
