# chatExtension
This is A Google Chrome browser extension that enables text chat among users that are watching the same YouTube video.

## How to run it locally

You will need:
- Docker, Docker-compose
- Node.js
- Google chrome browser



### Client:
1. From `client/` issue `npm run build`, which will create a `extension/build` directory. 
2. Navigate to the address `chrome://extensions` and click "load unpacked" in the top right corner, select to load `extension/build/build`
3. Now try to launch the extension. You might see an empty white box. If you do, continue following the steps. 
4. Navigate again to `chrome://extensions` and under the newly added extension, click "Errors". You should see something like the following:
  ```
  Refused to execute inline script because it violates the following Content
  Security Policy directive: "script-src 'self' 'sha256-ZQr+9PmEzw3a1P4rmCepcl4B2hHH7c9W8k/hpDQ+AaU='". 
  Either the 'unsafe-inline' keyword, a hash ('sha256-a7QNwV+1osAC6bXKzzgKk6A6flFr6M/rAKUL6EiBwzM='), or 
  a nonce ('nonce-...') is required to enable inline execution.
  ```
  Copy the new sha hash from it (e.g. `sha256-ZQr+9PmEzw3a1P4rmCepcl4B2hHH7c9W8k/hpDQ+AaU=`) and replace the old one at `content_security_policy` in `extensions/build/build/static/manifest.json`. 

5. Now refresh the extension from `chrome://extensions` and it should work. If you open a YouTube video and then the extension, you will see the interface.

### Server:

If you want to run the multiserver version of the application, stay on the main branch, otherwise check out the singleserver branch. 

1. Issue `docker build -t wsapp:latest ./server/` to build an image of the server
2. Issue `docker build -t revproxy:latest ./reverseproxy` to build an image of the Nginx reverse proxy
3. Issue `docker-compose up` to launch a number of application instances, the Nginx reverse proxy, and the redis instance
4. One can add more application instances by simply adding more 
```
wsn:
    image: wsapp:latest
    environment:
      - APPID=N
    depends_on:
      - rds
```
to the `docker-compose.yaml` file. Here, the `APPID` environment variable is simply used to identify the server in the log output. 

Now you can see on the console that we can open multiple tabs and the reverse proxy will randomly assign us to an upstream websocket server. If two tabs are on the same video, both will get all messages sent within the channel of that video. 


![Sign in](https://github.com/miteshkumar77/chatExtension/blob/main/cap1.jpg?raw=true)

![Chat](https://github.com/miteshkumar77/chatExtension/blob/main/cap2.png?raw=true)


## How it works

### Concurrency
We have implemented a concurrent blocking queue to create a worker pool and a concurrent read optimized hashmap to store current rooms, users, and frequency data for rate limiting. The hashmap is implemented using sharding and `sync.RWMutex`. The queue is implemented by using go channels as locks. 

### Redis pub/sub
All `wsapp:latest` instances subscribe and publish to the same redis pub/sub channel. When an instance recieves a message from its websocket connection, it publishes the message to the common channel. Simultaneously, in another goroutine, the instance listens to the subscription and adds incoming messages to a `Jobs` concurrent queue. Multiple worker goroutines wait until the `Jobs` queue has a message in it, and then broadcast that message to all of the users connected to the particular room. In this way, all users in a particular room will recieve the message no matter what instance they are connected to, and we can scale to a much higher amount of websocket connections. 

### Rate limiting
We use a concurrent hashmap in order to store how many messages are currently in the `Jobs` queue that haven't yet been processed. If this number reaches higher than a constant set threshold, the user's websocket is disconnected. This prevents someone from running a script to repeatedly send messages and put unnecessary load on the server.

This project does not persist any data, as the goal was not to produce a long term chat that users can come back to. The server simply directs messages to the correct recipients. 
  



