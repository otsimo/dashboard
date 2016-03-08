FROM centurylink/ca-certs
MAINTAINER Sercan Degirmenci <sercan@otsimo.com>

ADD bin/otsimo-dashboard-linux-amd64 /opt/otsimo-dashboard/bin/otsimo-catalog

EXPOSE 18857

# enable verbose debug for now
CMD ["/opt/otsimo-dashboard/bin/otsimo-dashboard","--debug","--storage","mongodb"]
