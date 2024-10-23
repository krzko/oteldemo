FROM golang:1.23 AS builder

WORKDIR /go/src/github.com/krzko/oteldemo

COPY . .

ARG BUILD_VERSION=0.0.0
ARG BUILD_DATE=1970-01-01T00:00:00Z
ARG COMMIT_ID=unknown

RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${BUILD_VERSION} -X main.date=${BUILD_DATE} -X main.commit=${COMMIT_ID}" \
    -o /usr/bin/oteldemo -v /go/src/github.com/krzko/oteldemo/cmd/oteldemo

FROM cgr.dev/chainguard/static:latest

COPY --from=builder /usr/bin/oteldemo /

ENTRYPOINT ["/oteldemo"]
