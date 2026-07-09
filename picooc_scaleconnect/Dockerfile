FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o scaleconnect .


FROM alpine:latest

COPY --from=builder /app/scaleconnect /app/scaleconnect

COPY run.sh /run.sh

RUN chmod +x /app/scaleconnect
RUN chmod +x /run.sh

CMD ["/run.sh"]