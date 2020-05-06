FROM golang:alpine as builder

WORKDIR /app
Add . /app/
RUN cd /app && go install ./cmd/channels


FROM alpine

WORKDIR /app
COPY example.config.json /app/config.json
COPY --from=app-builder /go/bin/channels /app/channels

ENTRYPOINT ["/app/channels"]
