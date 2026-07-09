FROM alpine:latest

COPY scaleconnect /app/scaleconnect
COPY run.sh /run.sh

RUN chmod +x /app/scaleconnect /run.sh
RUN ls -l /app
RUN file /app/scaleconnect

CMD ["/run.sh"]