import './App.css';
import { useState } from 'react';
import NotYoutube from './views/NotYoutube';
import Chat from './views/Chat';
const youtube_watch_hostname = 'https://www.youtube.com/watch';
/*global chrome*/

function isUrlYoutubeVideo(url) {
  return url.substring(0, youtube_watch_hostname.length) === youtube_watch_hostname;
}

function getVideoID(url) {
  var video_id = url.split('v=')[1];
  var ampersandPosition = video_id.indexOf('&');
  if (ampersandPosition != -1) {
    video_id = video_id.substring(0, ampersandPosition);
  }
  return video_id
}

function App() {

  const activeTabIsYouTube = true;
  const currentUrl = "https://www.youtube.com/watch?v=oqVMTwH_PMY";
  // const [currentUrl, changeCurrentUrl] = useState("");
  // const [activeTabIsYouTube, changeActiveTabIsYoutube] = useState(false);

  // chrome.tabs.query({ active: true }, (tab_arr) => {
  //   changeCurrentUrl(tab_arr[0].url);
  //   changeActiveTabIsYoutube(isUrlYoutubeVideo(tab_arr[0].url));
  // });

  if (activeTabIsYouTube) {
    return <Chat videoID={getVideoID(currentUrl)} />
  } else {
    return <NotYoutube url={currentUrl} />
  }
}

export default App;