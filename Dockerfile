FROM golang:1.18 AS builder
WORKDIR /go/src 
ENV GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=0   
ADD . .
RUN go mod tidy -compat=1.18 && go build -tags netgo -o /bin/chirpstack-event-forward cmd/chirpstack-event-forward/main.go

FROM alpine:3.10
COPY --from=builder /bin/chirpstack-event-forward /
ENTRYPOINT ["/chirpstack-event-forward"]