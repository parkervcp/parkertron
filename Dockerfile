# ----------------------------------
# Pterodactyl Panel Dockerfile
# ----------------------------------

FROM golang:1.9-alpine

COPY . ./src

WORKDIR /go/src/parkertron

RUN apk add --no-cache --update go git curl lua-stdlib lua musl-dev g++ libc-dev tesseract-ocr tesseract-ocr-dev \
 && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
 && dep ensure

CMD [ "go", "run", "*.go"]