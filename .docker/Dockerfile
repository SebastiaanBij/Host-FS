# Global Arguments
#----------------------------------------------------------------------------------------------------------------------#
ARG GOLANG_VERSION="1.26"
ARG ALPINE_VERSION="3.23"

# Builder
#----------------------------------------------------------------------------------------------------------------------#
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /build
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o ./bin/host-fs ./src

# Runner
#----------------------------------------------------------------------------------------------------------------------#
FROM alpine:${ALPINE_VERSION}

COPY --from=builder /build/bin/host-fs /usr/local/bin/host-fs

ENTRYPOINT ["/usr/local/bin/host-fs"]
