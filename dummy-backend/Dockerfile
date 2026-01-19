FROM golang:1.26-rc-trixie

WORKDIR /app

COPY main.go .

RUN go build -o backend main.go

CMD ["./backend"]
