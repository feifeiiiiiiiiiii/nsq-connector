TAG?=latest
.PHONY: build

build:
	docker build -t feifeiiiiiiiiiii/nsq-connector:$(TAG) .
push:
	docker push feifeiiiiiiiiiii/nsq-connector:$(TAG)
