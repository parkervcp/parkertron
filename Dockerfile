# ----------------------------------
# Pterodactyl Panel Dockerfile
# ----------------------------------

FROM golang:1.9-alpine

WORKDIR /srv/parkertron

RUN apk add --no-cache --update git \
 && go get github.com/bwmarrin/discordgo \
 && go get github.com/sirupsen/logrus

COPY . ./

CMD [ "go", "run", "*.go"]