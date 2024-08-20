FROM golang:1.22.5

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/gophermart ./cmd/gophermart
COPY ./internal ./internal
COPY ./migrations ./migrations
COPY ./pkg ./pkg
COPY ./atlas.hcl ./atlas.hcl

RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/gophermart -o gophermart

CMD ["/app/cmd/gophermart/gophermart"]
