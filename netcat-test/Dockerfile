FROM alpine:latest

RUN apk add netcat-openbsd

COPY netcat-test.sh /netcat-test.sh

ENTRYPOINT [ "/bin/sh", "/netcat-test.sh" ]