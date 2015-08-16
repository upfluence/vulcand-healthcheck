FROM busybox:latest
MAINTAINER Alexis Montagne <alexis.montagne@gmail.com>

COPY etcdenv /etcdenv
COPY vulcand-healthcheck /vulcand-healthcheck

RUN chmod +x /etcdenv
RUN chmod +x /vulcand-healthcheck

ENV NAMESPACE /environments/global
ENV ETCD_URL http://172.17.42.1:4001

ENTRYPOINT /etcdenv -n ${NAMESPACE} -s http://172.17.42.1:4001 /vulcand-healthcheck
