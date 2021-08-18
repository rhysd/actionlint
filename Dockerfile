ARG GOLANG_VER=latest
ARG ALPINE_VER=latest

FROM golang:${GOLANG_VER} as builder
WORKDIR /go/src/app
COPY go.* *.go ./
COPY cmd cmd/
RUN go build ./cmd/actionlint

FROM alpine:${ALPINE_VER}
COPY --from=builder /go/src/app/actionlint /usr/local/bin/
RUN apk add shellcheck py3-pyflakes
USER guest
ENTRYPOINT ["/usr/local/bin/actionlint"]
