
# Image URL to use all building/pushing image targets
IMG ?= quay.io/awesomenix/repairman-manager:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
GOBIN ?= $(PWD)/bin
PLATFORM := $(shell go env GOOS;)
ARCH := $(shell go env GOARCH;)
HAS_KUBEBUILDER := $(shell command -v $(GOBIN)/kubebuilder;)
KUBEBUILDER_VERSION := 2.0.1

all: manager

# Run tests
test: generate fmt vet manifests
	KUBEBUILDER_ASSETS=$(GOBIN) go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml
	rm -f ./config/default/manager_image_patch.yaml-e

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifndef HAS_KUBEBUILDER
	curl -L --fail -O \
		"https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz" && \
		tar -zxvf kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz && \
		rm kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz && \
		mv ./kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH)/bin/* $(GOBIN) && \
		rm -rf ./kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH)
endif
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0
	
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
