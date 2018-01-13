# ----------------------------------
# Pterodactyl Panel Dockerfile
# ----------------------------------

FROM golang:1.9-alpine

WORKDIR /srv/parkertron

RUN apk add --no-cache --update git lua-stdlib lua musl-dev g++ libc-dev \
 && go get github.com/bwmarrin/discordgo \
 && go get github.com/sirupsen/logrus \
 && go get github.com/otiai10/gosseract \
 && mvdan.cc/xurls

COPY . ./

CMD [ "go", "run", "*.go"]