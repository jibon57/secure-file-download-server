FROM golang:1.25 as builder

ARG TARGETPLATFORM
ARG TARGETARCH
RUN echo building for "$TARGETPLATFORM"

WORKDIR /go/src/app

COPY go.mod go.mod
COPY go.sum go.sum
# download if above files changed
RUN go mod download

# Copy the go source
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags '-w -s -buildid=' -a -o download-server

FROM alpine

COPY --from=builder /go/src/app/download-server /usr/bin/download-server

# Run the binary.
ENTRYPOINT ["download-server"]
