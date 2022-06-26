package config

type Config struct {
	PortsFileLocation string `envconfig:"PORTS_FILE" default:"./helperFiles/portsSmall.json"`
	LOG               LogConfig
	Elastic           ElasticConfig
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
