# Copyright (c) 2019 StackRox Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Makefile for building the Admission Controller webhook demo server + docker image.

.DEFAULT_GOAL := docker-image

K8S_NAMESPACE := webhook-demo
KUBECTL_EXE := /usr/bin/kubectl
K8S_CMD := $(KUBECTL_EXE) -n $(K8S_NAMESPACE)
DOCKER_ACCOUNT = rtisma1
IMAGE ?= $(DOCKER_ACCOUNT)/webhook-go-server:latest

# this is the legacy build that is also replicated in the Dockerfile
image/webhook-server: $(shell find . -name '*.go')
	./init-build.sh
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/webhook-server

docker-image:
	docker build -t $(IMAGE) ./

push-image: docker-image
	docker push $(IMAGE)

send-request:
	cd ./examples/ && ./run-server-request.sh 8080

start-docker:
	docker-compose up --build -d

stop-docker:
	docker-compose down -v

k8s-destroy:
	@./destroy.sh

k8s-deploy:
	@./deploy.sh

k8s-test:
	@$(K8S_CMD) apply -f ./examples/pod-with-defaults.yaml
	@$(K8S_CMD) get pods -oyaml pod-with-defaults
