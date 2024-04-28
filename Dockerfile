FROM golang:1.22.2-alpine3.19

WORKDIR /goproxy

COPY go.mod ./

RUN go mod download && go mod verify

COPY ./ ./

RUN go build -v -o /usr/local/bin/goproxy

EXPOSE 80

CMD ["goproxy"]
