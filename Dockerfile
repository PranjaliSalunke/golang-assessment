
FROM golang:latest

WORKDIR /go/src/app

COPY . .

RUN go get -u github.com/gorilla/mux

RUN go install -v ./...

EXPOSE 8000

CMD ["myproject"]
