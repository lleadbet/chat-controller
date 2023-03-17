# chat-controller

A simple golang executable for use with Twitch to allow chat to input aribtrary buttons. 

## How to use

Firstly, download the most recent version from the Releases tab, on the right, matching your OS version. 

Next, create a `config.yaml` file in the same folder you've downloaded the exe into. This file will look like: 

```yaml
username: concreteentree
chat_message:
  - key: W
    duration: 2
    message: 
    - forward
    - up
    - go
    - move
  - key: D
    duration: 0.2
    message: left
  - key: A
    duration: 0.2
    message: right
  - key: S
    duration: 0.2
    message: 
    - reverse
    - back
    - retreat
```

The configuration has the following options: 

### Root-level

`username` - the username of the channel to join to listen to for events. If you're unsure what this is, go to your Twitch channel page and it'll be text after https://twitch.tv/

`chat_messages` - A sequence of values to assign to various chat messages. 

#### Chat Messages

Each entry consists of: 

`key` - the key(s) to be pushed on the chat command. When multiple have been entered, all will be exceuted. These _must_ match 

`duration` - the duration the keys will be pressed for, in seconds. 

`message` - the chat messages (case insensitive) to respond to. The messages must match exactly minus whitespace and casing (meaning a value of `pog` would not trigger on messages like `Poggers` but `Pog` would)


