# parkertron

A simple discord chat bot with a simple configuration. Written using [discordgo](https://github.com/bwmarrin/discordgo).

Requirements:


Working on adding other services and additions.

The checklist so far

- [x] Full OSS release

- [x] Read config files
  - [ ] Separate service config

- [ ] support multiple services
  - [x] Discord
  - [ ] Slack
  - [ ] IRC

- [x] Get inbound messages
  - [x] Listen to specific channels
  - [ ] Listen for mentions


- [ ] Respond to inbound messages
  - [ ] respond to commands with prefix
  - [ ] respond to key words/phrases
  - [ ] Comma separated word lists
  - [ ] Separate server commands
- [x] Image parsing
  - [x] image from url support
    - [x] png support
    - [x] jpg support
  - [ ] direct embedded images

- [ ] Respond with correct output from configs

- [ ] Respond with multi-line output in a single message

- [ ] Impliment blacklist/whitelist mode

- [ ] Mitigate spam with cooldown per user/channel/global
  - [ ] global cooldown
  - [ ] channel cooldown
  - [ ] user cooldown

- [ ] Permissions
  - [ ] Permissions management

- [ ] Logging
  - [ ] Log user join/leave 
  - [ ] Log chats
  - [ ] Log edits (show original and edited)
  - [ ] Log chats/edits to separate files/folders
  
- [ ] Join voice channels
  - [ ] Play audio from links


So far I have the chat bot part down with no limiting or administration.

Configuration is done in yaml/json.  
If you have a Bot account already you can add the token and client ID's on your own.  
If you don't you will need to set your own account up.

The "owner" option in the configs is basically a super admin that will not be able to be blacklisted.

The prefix is the command prefix and is customizable.  
Set to "." by default it can be changed to whatever you want.


The Commands set up is simple and is also in json.  
See the commands.json for examples.  
