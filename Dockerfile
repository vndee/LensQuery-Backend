FROM golang:latest

LABEL maintainer="Duy Huynh <vndee.huynh@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/app/main.go

EXPOSE 8080

CMD ["./main", "--name=Lensquery-Backend", "--port=8080", "--prod"]
