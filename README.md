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

![Sign in](https://github.com/miteshkumar77/chatExtension/blob/main/cap1.jpg?raw=true)

### Server:

If you want to run the multiserver version of the application, stay on the main branch, otherwise check out the singleserver branch. Regardless, the rest of the steps are the same. 

1. Issue `docker build -t wsapp:latest ./server/` to build an image of the server
2. Issue `docker build -t revproxy:latest ./reverseproxy` to build an image of the Nginx reverse proxy
3. Issue `docker-compose up` to launch a number of application instances, the Nginx reverse proxy, and the redis instance. Now you will be able to use the interface to connect to the chat room of the YouTube video that is open in your active tab. 

![Chat](https://github.com/miteshkumar77/chatExtension/blob/main/cap2.png?raw=true)


4. One can add more application instances by simply adding more of the following:
```
# docker-compose.yaml
ws< N >:
    image: wsapp:latest
    environment:
      - APPID=< N >
    depends_on:
      - rds
     
     
# reverseproxy/templates/nginx.tmpl
upstream ws-backend {
         .
         .
         .
  server ws< N >:8080;
}

```
Here, the `APPID` environment variable is simply used to identify the server in the log output. 

Now you can see on the console that we can open multiple tabs and the reverse proxy will randomly assign us to an upstream websocket server. If two tabs are on the same video, both will get all messages sent within the chat room of that video. 



## How it works

### Concurrency
We have implemented a concurrent blocking queue to create a task queue that is consumed by a worker pool. The queue is implemented using channels as locks so that waiting threads can stall until a new task is available without having to poll the queue themselves. 
We have also implemented a concurrent read optimized hashmap to store concurrent rooms, users, and frequency data for rate limiting. The hashmap is implemented using sharding and `sync.RWMutex`. It also has a good API for doing updates based on original values of the map while avoiding race conditions.

### Redis pub/sub
All `wsapp:latest` instances subscribe and publish to the same redis pub/sub channel. When an instance recieves a message from its websocket connection, it publishes the message to the common channel. Simultaneously, in another goroutine, the instance is continuously listening to the subscription and adding incoming messages to a `Jobs` concurrent queue. Multiple worker goroutines pull messages off of the `Jobs` queue and broadcast them to all of the users connected to the particular room associated with the message sender. In this way, all users in a particular room will recieve the message no matter what instance they are connected to, and we can scale to a much higher amount of websocket connections. 

### Rate limiting
We use a concurrent hashmap in order to store how many messages are currently in the `Jobs` queue that haven't yet been processed. If this number reaches higher than a constant set threshold, the user's websocket is disconnected. This prevents someone from running a script to repeatedly send messages and put unnecessary load on the server.

This project does not persist any data, as the goal was not to produce a long term chat that users can come back to. The server simply directs messages to the correct recipients. 
  



