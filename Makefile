all: push

# 0.0 shouldn't clobber any released builds
TAG = 0.0
PREFIX = aledbf/gobgp

binaries:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' -o ./docker/gobgpd ./gobgpd
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' -o ./docker/gobgp ./gobgp

container: binaries
	docker build -t $(PREFIX):$(TAG) docker

push: container
	docker push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
	rm -rf ./docker/gobgpd ./docker/gobgp
