FROM golang:1.16-alpine AS builder
WORKDIR /go/src/github.com/christian-korneck/docker-pushrm
COPY . .
RUN apk add --no-cache ca-certificates && update-ca-certificates
ENV CGO_ENABLED 0
RUN go build -v -a -tags netgo -ldflags='-s -w -extldflags "-static"' .
#RUN apk add --no-cache upx && upx ./docker-pushrm

FROM docker.io/library/busybox
COPY --from=builder /go/src/github.com/christian-korneck/docker-pushrm/docker-pushrm /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#nobody:nogroup
USER 65534:65534
ENTRYPOINT ["/docker-pushrm"]
