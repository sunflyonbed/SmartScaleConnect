FROM alpine:latest

COPY scaleconnect /usr/bin/scaleconnect
COPY run.sh /run.sh

RUN chmod +x /usr/bin/scaleconnect
RUN chmod +x /run.sh

CMD ["/run.sh"]