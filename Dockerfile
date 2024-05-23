FROM golang:1.22-alpine
WORKDIR /usr/src/app

# Add dependencies
RUN apk update && \
    apk add netcat-openbsd && \
    apk add sudo

ENTRYPOINT /bin/sh
