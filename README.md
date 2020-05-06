# channels

[![Docker pulls](https://img.shields.io/docker/pulls/everpcpc/channels.svg)](https://hub.docker.com/r/everpcpc/channels)
[![Go Report Card](https://goreportcard.com/badge/github.com/everpcpc/channels)](https://goreportcard.com/report/github.com/everpcpc/channels)


channels is a message gateway by channel.

With a stateless irc protocol subset server and streaming web api endpoints.

And backend message queue based on redis pubsub.

## TODO

- [x] irc auth with ldap
- [x] api auth with tokens
- [ ] api auth with openid
- [ ] role based channel permission
- [ ] github webhook support
- [ ] sentry webhook support
- [ ] forwarder component to work as a message relay
- [ ] sse/websocket api endpoint
- [ ] web ui
- [ ] mysql backend for persistent store
