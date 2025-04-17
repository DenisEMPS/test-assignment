FROM golang:1.24.0

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o App cmd/app/main.go

CMD [ "./App" ]