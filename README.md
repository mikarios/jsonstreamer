# jsonstreamer
The purpose of jsonstreamer is to parse a big json file with specific structure into a database.

An example of such a file can be found inside [third party folder](third_party).

The process starts from cmd/main.go

* For demonstration purposes the data is inserted into an elasticsearch.

For convenience a Makefile is included with the following commands:
- `start-elastic` 
- `rm-elastic`
- `lint`
- `test`
- `restart-elastic-and-test`
- `run`

`make start-elastic` and `make rm-elastic` start and remove a docker with elastic 7.17.4

`make test` tests against this elastic

`restart-elastic-and-test` well, it does what is in the name

In order for jsonstreamer to execute the following environment variables are needed (default values as stated)
- LOG_FORMAT=text
- LOG_LEVEL=debug
- LOG_TRACE=true
- PORTS_FILE=./third_party/ports.json
- MAX_MEMORY_MB=0
- PORT_COLLECTOR_WORKERS=0
- ELASTIC_URLS=http://localhost:9200
- ELASTIC_IDX_PORTS_REPLICAS=0
- ELASTIC_IDX_PORTS=ports

For the end to end test to execute these values are changed with the important one being ELASTIC_IDX_PORTS which takes the value test-ports

A simple `docker-compose up -d && docker logs -f jsonparser` should execute the jsonparser against the given [ports.json](third_party/ports.json)

### To improve:
* Better distinction between database service and portCollector service. Didn't have enough time, and it seemed like over engineering for this specific task since there was no manipulation of port data needed.
  * However, better distinction would allow us to send data to elastic in batches and not one by one, making it a better practice for our database
* More tests needed but time was limited