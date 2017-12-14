# parkertron

A simple discord chat bot with a simple configuration. Written using

[discordgo](https://github.com/bwmarrin/discordgo)  


Working on adding other services and additions.

- [x] Full OSS release

###### Console Requirements
- [ ] Read config files
- [ ] Console commands
- [ ] Logging
- [ ] Respond with correct output from configs
- [ ] Respond with multi-line output in a single message
- [ ] multithreading (goroutines)

###### Discord specific
- [ ] Get bot connected to discord
  - [ ] Per server configs
  - [ ] Multiple bots on different threads
- [ ] Get inbound messages
  - [ ] Listen to specific channels
  - [ ] Listen for mentions
  - [ ] Listen for command prefix
  - [ ] Listen for key words/phrases
- [ ] Respond to inbound messages
  - [ ] Respond correctly
  - [ ] Comma separated word lists
- [ ] Impliment blacklist/whitelist mode
- [ ] Mitigate spam with cooldown per user/channel/global
  - [ ] global cooldown
  - [ ] channel cooldown
  - [ ] user cooldown
- [ ] Impliment permissions
- [ ] Impliment permissions management
- [ ] Server logging
  - [ ] Log user join/leave
  - [ ] Log chats
  - [ ] Log edits (show original and edited)
  - [ ] Log chats/edits to separate files/folders
- [ ] Join voice channels
  - [ ] Play audio from links

###### Other Services

Configuration is done in json.  
If you have a Bot account already you can add the token and client ID's on your own.  
If you don't you will need to set your own account up.

The prefix is the command prefix and is customizable.  
Set to "." by default it can be changed to whatever you want.
