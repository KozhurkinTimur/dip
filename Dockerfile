FROM alpinelinux/golang:3.20

WORKDIR /app

COPY ./go.mod /app/go.mod
RUN go mod download

COPY . /app

CMD ["docker", "compose", "up"]
CMD ["go", "run", "main.go"]
