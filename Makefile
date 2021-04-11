.PHONY: all
all:
	go build -o _output/bind-host ./cmd/bind

.PHONY: in image
in image:
	kubectl dev build --local _output/

.PHONY: image
image:
	kubectl dev build -t docker.io/warmmetal/bind-host:latest