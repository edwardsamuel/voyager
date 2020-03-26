
all: install

proto_all: lint generate

lint:
	prototool lint

generate:
	prototool generate

install:
	go install github.com/vvarma/voyager/cmd/voyager

docker:
	docker build . -f build/Dockerfile -t voyager
