package config

type Config struct {
	PortsFileLocation    string `envconfig:"PORTS_FILE" default:"./third_party/ports.json"`
	MaxMemoryAvailable   uint64 `envconfig:"MAX_MEMORY_MB"`
	PortCollectorWorkers int    `envconfig:"PORT_COLLECTOR_WORKERS" default:"2"`
	LOG                  LogConfig
	Elastic              ElasticConfig
}

type ElasticConfig struct {
	URLList []string `envconfig:"ELASTIC_URLS" default:"http://localhost:9200"`
	Indices struct {
		Ports struct {
			Replicas int    `envconfig:"ELASTIC_IDX_PORTS_REPLICAS"`
			Index    string `envconfig:"ELASTIC_IDX_PORTS" default:"ports"`
		}
	}
}

type LogConfig struct {
	Format string `envconfig:"LOG_FORMAT" default:"text"`
	Level  string `envconfig:"LOG_LEVEL" default:"debug"`
	Trace  bool   `envconfig:"LOG_TRACE" default:"true"`
}
