FROM centurylink/ca-certs
MAINTAINER Sercan Degirmenci <sercan@otsimo.com>

ADD dashboard-linux-amd64 /opt/otsimo/dashboard
COPY sample.yaml /opt/otsimo/config.yaml

EXPOSE 18860
CMD ["/opt/otsimo/dashboard","--debug","--storage","mongodb"]
