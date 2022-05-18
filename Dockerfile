FROM golang:1.18-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/infinimesh/http-fs
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -buildvcs=false
RUN upx ./http-fs

FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/infinimesh/http-fs/http-fs /http-fs

LABEL org.opencontainers.image.source https://github.com/infinimesh/http-fs

ENTRYPOINT ["/http-fs"]
