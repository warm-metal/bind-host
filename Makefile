.PHONY: all
all:
	go build -o _output/bind-host ./cmd/bind

.PHONY: in container
in container:
	kubectl dev build --local _output/

.PHONY: image
image:
	kubectl dev build -t docker.io/warmmetal/bind-host:latest

.PHONY: test
test:
	kubectl dev build -t docker.io/warmmetal/bind-host-test:integration test
	kubectl delete --ignore-not-found -f test/manifest.yaml
	kubectl apply --wait -f test/manifest.yaml
	kubectl wait --timeout=1m --for=condition=complete job/bind-host-test