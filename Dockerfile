FROM golang:1.24 AS builder

WORKDIR /app

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o flight-booking .

FROM scratch

WORKDIR /app

COPY ./public ./public
COPY --from=builder /app/flight-booking ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 80
ENTRYPOINT ["./flowers"]