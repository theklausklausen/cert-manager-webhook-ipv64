GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

IMAGE_NAME := "repo.kk.int/cert-manager-webhook-ipv64"
IMAGE_TAG := "latest"

OUT := $(shell pwd)/_out

KUBEBUILDER_VERSION=1.28.0

HELM_FILES := $(shell find charts/cert-manager-webhook-ipv64)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

test: _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/etcd _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kube-apiserver _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kubectl
	TEST_ASSET_ETCD=_test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/etcd \
	TEST_ASSET_KUBE_APISERVER=_test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kube-apiserver \
	TEST_ASSET_KUBECTL=_test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kubectl \
	$(GO) test -v .

_test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH).tar.gz: | _test
	curl -fsSL https://go.kubebuilder.io/test-tools/$(KUBEBUILDER_VERSION)/$(OS)/$(ARCH) -o $@

_test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/etcd _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kube-apiserver _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)/kubectl: _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH).tar.gz | _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH)
	tar xfO $< kubebuilder/bin/$(notdir $@) > $@ && chmod +x $@

.PHONY: clean
clean:
	rm -r _test $(OUT)

.PHONY: build
build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml: $(OUT)/rendered-manifest.yaml

$(OUT)/rendered-manifest.yaml: $(HELM_FILES) | $(OUT)
	helm template \
	    --name-template cert-manager-webhook-ipv64 \
            --set image.repository=$(IMAGE_NAME) \
            --set image.tag=$(IMAGE_TAG) \
            charts/cert-manager-webhook-ipv64 > $@

_test $(OUT) _test/kubebuilder-$(KUBEBUILDER_VERSION)-$(OS)-$(ARCH):
	mkdir -p $@

run:
	go run main.go

link-tc:
	rm -rf /home/$(USER)/.config/Code/User/globalStorage/rangav.vscode-thunder-client
	ln -s ${PWD}/.vscode/rangav.vscode-thunder-client /home/$(USER)/.config/Code/User/globalStorage/

test-suite:
	docker run --rm -v $(PWD):/src -w /src golang:1.23.4 sh -c 'cd /src && TEST_ZONE_NAME=example.com make test'

go-mod-tidy:
	docker run --rm -v $(PWD):/src -w /src golang:1.23.4 sh -c 'cd /src && go mod tidy'