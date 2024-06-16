FROM golang:alpine

COPY / /goruut
WORKDIR /goruut/cmd/goruut
RUN go mod tidy
RUN go install
WORKDIR /goruut
ENTRYPOINT ["/go/bin/cmd", "--configfile", "/goruut/configs/config.json"]
