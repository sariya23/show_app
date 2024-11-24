# docker run --net=host -it --rm server:v1

FROM golang:1.23.1-alpine3.19
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main cmd/main.go
EXPOSE 8082
RUN apk update