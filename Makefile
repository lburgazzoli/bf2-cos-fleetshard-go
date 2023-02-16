
# Image URL to use all building/pushing image targets
IMG ?= quay.io/lburgazzoli/cos-fleetshard:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCAL_BIN_PATH := ${PROJECT_PATH}/bin
KO_CONFIG_PATH := ${PROJECT_PATH}/etc/ko.yaml
KO_DOCKER_REPO := "quay.io/lburgazzoli/cos-fleetshard"
CGO_ENABLED := 0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen
	#$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...
#test: manifests generate fmt vet envtest ## Run tests.
#	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: test/integration
test/integration:
	go test -v ./internal/camel/test -tags=integration

##@ Build

.PHONY: build
build: manifests generate fmt vet
	CGO_ENABLED=0 go build -o bin/fleetshard cmd/fleetshard/main.go

.PHONY: run
run: manifests generate fmt vet
	CGO_ENABLED=0 go run cmd/fleetshard/main.go run \
		--leader-election=false \
		--operator-id foo \
		--operator-group cos.bf2.dev \
		--operator-type camel \
		--pprof-bind-address localhost:6060 \
		--metrics-bind-address localhost:8080 \
		--health-probe-bind-address localhost:8081


.PHONY: openapi/generate
generate/openapi: openapi-generator
	$(OPENAPI_GENERATOR) validate -i etc/openapi/connector_mgmt-private.yaml
	$(OPENAPI_GENERATOR) generate -i etc/openapi/connector_mgmt-private.yaml \
		-g go \
		-o internal/api/controlplane \
		--package-name controlplane  \
		--additional-properties=enumClassPrefix=true \
		--additional-properties=isGoSubmodule=false \
		--additional-properties=withGoCodegenComment=true \
		--ignore-file-override etc/openapi/.openapi-generator-ignore



	rm internal/api/controlplane/go.sum
	rm internal/api/controlplane/go.mod
	rm internal/api/controlplane/git_push.sh
	rm internal/api/controlplane/.travis.yml
	rm internal/api/controlplane/.openapi-generator-ignore
	rm internal/api/controlplane/README.md

	rm -rf internal/api/controlplane/docs
	rm -rf internal/api/controlplane/api
	rm -rf internal/api/controlplane/.openapi-generator
	rm -rf internal/api/controlplane/test

	$(GOFMT) -w internal/api/controlplane
	$(GOIMPORT) -w internal/api/controlplane


##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: image/publish
image/publish: ko
	KO_CONFIG_PATH=$(KO_CONFIG_PATH) KO_DOCKER_REPO=${KO_DOCKER_REPO} $(KO) build --sbom none --bare ./cmd/fleetshard

.PHONY: image/local
image/local: ko
	KO_CONFIG_PATH=$(KO_CONFIG_PATH) KO_DOCKER_REPO=ko.local $(KO) build --sbom none --bare ./cmd/fleetshard

.PHONY: image/kind
image/kind: ko
	KO_CONFIG_PATH=$(KO_CONFIG_PATH) KO_DOCKER_REPO=kind.local $(KO) build --sbom none --bare ./cmd/fleetshard

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOIMPORT ?= $(LOCALBIN)/goimports
KO ?= $(LOCALBIN)/ko

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.0
CONTROLLER_TOOLS_VERSION ?= v0.11.3

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)


.PHONY: goimport
goimport: $(GOIMPORT)
$(GOIMPORT): $(LOCALBIN)
	test -s $(LOCALBIN)/goimport || \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest

.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: controller-gen
ko: $(KO)
$(KO): $(LOCALBIN)
	test -s $(LOCALBIN)/ko || GOBIN=$(LOCALBIN) go install github.com/google/ko@main

OPENAPI_GENERATOR ?= ${LOCAL_BIN_PATH}/openapi-generator
NPM ?= "$(shell which npm)"
openapi-generator:
ifeq (, $(shell which ${NPM} 2> /dev/null))
	@echo "npm is not available please install it to be able to install openapi-generator"
	exit 1
endif
ifeq (, $(shell which ${LOCAL_BIN_PATH}/openapi-generator 2> /dev/null))
	@{ \
	set -e ;\
	mkdir -p ${LOCAL_BIN_PATH} ;\
	mkdir -p ${LOCAL_BIN_PATH}/openapi-generator-installation ;\
	cd ${LOCAL_BIN_PATH} ;\
	${NPM} install --prefix ${LOCAL_BIN_PATH}/openapi-generator-installation @openapitools/openapi-generator-cli ;\
	ln -s openapi-generator-installation/node_modules/.bin/openapi-generator-cli openapi-generator ;\
	}
endif


.PHONY: openapi-generator