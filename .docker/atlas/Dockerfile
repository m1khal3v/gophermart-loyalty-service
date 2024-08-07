FROM arigaio/atlas:0.25.0-alpine

WORKDIR /migrations

COPY --from=golang:1.22.5-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/gophermart ./cmd/gophermart
COPY ./internal ./internal
COPY ./migrations ./migrations
COPY ./pkg ./pkg
COPY ./atlas.hcl ./atlas.hcl

ENTRYPOINT ["atlas"]
