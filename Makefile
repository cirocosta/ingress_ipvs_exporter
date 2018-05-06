VERSION             := $(shell cat ./VERSION)
COMMIT_SHA          := $(shell git rev-parse --short HEAD)
DOCKER_FINAL_IMAGE  := cirocosta/ingress_ipvs_exporter

all: install


install:
	go install -v


test:
	go test -v ./...


fmt:
	find ./mapper -name "*.c" -o -name "*.h" | \
		xargs clang-format -style=file -i
	go fmt ./...


mapper/libmapper.so: ./mapper/mapper.c ./mapper/mapper.h
	gcc $< \
		-fPIC \
		-shared \
		-c \
		-o $@


mapper.out: ./mapper/cmd/main.c mapper/libmapper.so
	gcc $< \
		-o $@ \
		-L./mapper \
		-lmapper \
		-lip4tc \
		-lxtables


image:
	docker build \
		-t $(DOCKER_FINAL_IMAGE):$(VERSION) \
		.
	docker tag \
		$(DOCKER_FINAL_IMAGE):$(VERSION) \
		$(DOCKER_FINAL_IMAGE):$(VERSION)-$(COMMIT_SHA)
	docker tag \
		$(DOCKER_FINAL_IMAGE):$(VERSION) \
		$(DOCKER_FINAL_IMAGE):latest


login:
	echo $(DOCKER_PASSWORD) | docker login \
		--username $(DOCKER_USERNAME) \
		--password-stdin


push: login
	docker push $(DOCKER_FINAL_IMAGE):$(VERSION)
	docker push $(DOCKER_FINAL_IMAGE):$(VERSION)-$(COMMIT_SHA)
	docker push $(DOCKER_FINAL_IMAGE):latest


