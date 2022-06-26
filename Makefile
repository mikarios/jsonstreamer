.PHONY: start-elastic lint

start-elastic:
	docker run -p 127.0.0.1:9200:9200 -p 127.0.0.1:9300:9300 -e "discovery.type=single-node" -d docker.elastic.co/elasticsearch/elasticsearch:7.17.4

lint:
	golangci-lint run -c .golangci.yml