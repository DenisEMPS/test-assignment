FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main cmd/app/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .

CMD [ "./main" ]