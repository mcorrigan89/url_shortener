# syntax=docker/dockerfile:1

FROM golang:1.22
WORKDIR /usr/src/app

COPY . .
RUN make ui
RUN make css
RUN go build -o=./bin/main ./cmd

EXPOSE 9001

CMD ["./bin/main"]
