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

IMAGE ?= rtisma1/webhook-go-server:latest

image/webhook-server: $(shell find . -name '*.go')
	./init-build.sh
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/webhook-server

.PHONY: dev-docker-image
dev-docker-image: image/webhook-server
	docker build -t $(IMAGE) image/

.PHONY: dev-push-image
dev-push-image: docker-image
	docker push $(IMAGE)

dev-run-server: image/webhook-server
	@echo "Running server"
	./image/webhook-server &

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
