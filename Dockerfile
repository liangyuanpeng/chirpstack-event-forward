FROM golang:1.24 AS builder
WORKDIR /go/src 
ENV GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=0   
ADD . .
RUN go mod tidy && go build -tags netgo -o /bin/chirpstack-event-forward cmd/chirpstack-event-forward/main.go

FROM alpine:3.22
COPY --from=builder /bin/chirpstack-event-forward /
ENTRYPOINT ["/chirpstack-event-forward"]
