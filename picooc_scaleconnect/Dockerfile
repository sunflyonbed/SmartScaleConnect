FROM alpine:latest

COPY scaleconnect /app/scaleconnect
COPY run.sh /run.sh

RUN chmod +x /app/scaleconnect /run.sh

VOLUME /config
WORKDIR /config

CMD ["/run.sh"]