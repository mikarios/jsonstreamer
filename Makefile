.PHONY: start-elastic lint test

start-elastic:
	docker run --name es-test -p 127.0.0.1:9200:9200 -p 127.0.0.1:9300:9300 -e "discovery.type=single-node" -d docker.elastic.co/elasticsearch/elasticsearch:7.17.4

rm-elastic:
	docker rm -f es-test

lint:
	golangci-lint run -c .golangci.yml

test: rm-elastic start-elastic
	sleep 10
	go test ./...