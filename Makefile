build:
	go build -v ./...

build-image:
	docker build -t ghcr.io/liangyuanpeng/chirpstack-event-forward .
