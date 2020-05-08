# channels

[![Docker pulls](https://img.shields.io/docker/pulls/everpcpc/channels.svg)](https://hub.docker.com/r/everpcpc/channels)
[![Go Report Card](https://goreportcard.com/badge/github.com/everpcpc/channels)](https://goreportcard.com/report/github.com/everpcpc/channels)


channels is a message hub/gateway by channel.

With a stateless irc protocol subset server and streaming web api endpoints.

And backend message queue based on redis pubsub.

### RUN

```shell
docker run --rm --net=host \
    -v /path/to/config.json:/app/config.json:ro \
    -it everpcpc/channels --help
```

### DONE

- [x] irc auth with ldap
- [x] api auth with tokens
- [x] sentry webhook support
- [x] github webhook support
- [x] alertmanager webhook support

### TODO

- [ ] api auth with openid
- [ ] role based channel permission
- [ ] forwarder component for working as a message relay
- [ ] sse/websocket api endpoint
- [ ] web ui
- [ ] mysql backend for persistent store
- [ ] kafka backend for group consuming
