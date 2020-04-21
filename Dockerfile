# Stage 1
FROM golang:1.13.9-alpine3.11 as builder

ARG BUILD_TOKEN

# Add git
RUN apk update && \
    apk add git && \
    apk add openssl-dev && \
    apk add gcc && \
    apk add libc-dev

RUN mkdir $GOPATH/src/github.com
RUN mkdir $GOPATH/src/github.com/YOVO-LABS
RUN mkdir $GOPATH/src/github.com/YOVO-LABS/k8scaler

ADD . $GOPATH/src/github.com/YOVO-LABS/k8scaler/
WORKDIR $GOPATH/src/github.com/YOVO-LABS/k8scaler

RUN export GOPRIVATE=github.com/YOVO-LABS/*

# RUN ls
RUN go get ./...
RUN go build .

# Stage 2
FROM alpine:3.11

RUN apk add --update \
    curl \
    ca-certificates

COPY --from=builder /go/bin/k8scaler /
EXPOSE 4000

CMD ["./k8scaler"]

# docker run -d --name transcoder --env-file=.env -p 4000:4000 asia.gcr.io/chrome-weft-229408/scaler:v1
