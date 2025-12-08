build:
	go build -o bin/mdserve main.go
	
run:
	./bin/mdserve

clean:
	rm -rf bin/
	rm -rf .static/
	rm -rf .git-remote-content/

build-docker:
	docker build -t mdserve .

build-and-run:
	make build
	make run