FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.work .
COPY shared/ shared/
COPY departures-api/ departures-api/
RUN cd departures-api && go build -o /out .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out /out
CMD ["/out"]
