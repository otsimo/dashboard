FROM alpine:3.3
MAINTAINER Sercan Degirmenci <sercan@otsimo.com>

RUN apk-install ca-certificates

ADD welcome-linux-amd64 /opt/otsimo/welcome
COPY templates /opt/otsimo/templates

EXPOSE 18864
CMD ["/opt/otsimo/welcome","--debug"]
