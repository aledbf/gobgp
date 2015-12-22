FROM alpine:3.2

RUN apk add -U wget curl tcpdump supervisor quagga bash && \
  rm -rf /var/cache/apk/*

RUN touch /etc/quagga/zebra.conf /etc/quagga/bgpd.conf

COPY supervisord.conf /etc/supervisord.conf
COPY gobgp /usr/sbin/gobgp
COPY gobgpd /usr/sbin/gobgpd

ENV HOME /root

WORKDIR /

VOLUME /etc/quagga

ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
