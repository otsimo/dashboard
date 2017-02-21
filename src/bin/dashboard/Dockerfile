FROM alpine:3.5
RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*

ADD dashboard-linux-amd64 /opt/otsimo/dashboard
COPY sample.yaml /opt/otsimo/config.yaml

EXPOSE 18860
ENTRYPOINT ["/opt/otsimo/dashboard"]
