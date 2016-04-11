FROM alpine:3.3
MAINTAINER Sercan Degirmenci <sercan@otsimo.com>

RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*

ADD welcome-linux-amd64 /opt/otsimo/welcome
COPY templates /opt/otsimo/templates

EXPOSE 18864
CMD ["/opt/otsimo/welcome","--debug"]
