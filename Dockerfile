# ----------------------------------
# parkertron dockerfile
# ----------------------------------

FROM golang:1.22-bookworm

COPY . /parkertron

WORKDIR /parkertron

RUN apt update -y \
 && apt install -y tesseract-ocr tesseract-ocr-eng libtesseract-dev \
 && go mod tidy \
 && go build -o parkertron

FROM debian:bookworm-slim

RUN apt update -y \
    && apt install -y iproute2 ca-certificates libtesseract-dev tesseract-ocr-eng

WORKDIR /app/

COPY --from=0 /parkertron/parkertron /app/

VOLUME /app/configs
VOLUME /app/logs

CMD ["./parkertron"]