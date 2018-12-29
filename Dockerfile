# ----------------------------------
# Pterodactyl Panel Dockerfile
# ----------------------------------

FROM golang:1.11-alpine

COPY . /parkertron

WORKDIR /parkertron

RUN apk add --no-cache --update git curl lua-stdlib lua musl-dev g++ libc-dev tesseract-ocr tesseract-ocr-dev \
 && go mod tidy \
 && go build

CMD ["/go/src/parkertron/parkertron"]