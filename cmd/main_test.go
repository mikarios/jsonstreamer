package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
)

var (
	testDataFile  = "../third_party/portsSmall.json"
	elasticURL    = "http://localhost:9200"
	testPortIndex = "test-ports"
)

// I don't have time to mock the db. Normally I would simply create a mock with a map in memory and use that for testing
func Test_EndToEnd(t *testing.T) {
	cfg := config.Init("")
	cfg.Elastic.URLList = []string{elasticURL}
	cfg.PortCollectorWorkers = 4
	cfg.PortsFileLocation = testDataFile
	cfg.MaxMemoryAvailable = 500
	cfg.Elastic.Indices.Ports.Index = testPortIndex
	cfg.Elastic.Indices.Ports.Replicas = 0

	notifyOnFinish := make(chan interface{})
	startServices(context.Background(), cfg, notifyOnFinish)

	select {
	case <-time.After(1 * time.Second):
		t.Error("did not receive finish signal")
		t.Fail()
	case <-notifyOnFinish:
	}

	var portsFromFile map[string]portmodel.PortData

	f, _ := os.Open(testDataFile)

	if err := json.NewDecoder(f).Decode(&portsFromFile); err != nil {
		t.Errorf("could not decode file: %v", err)
		t.Fail()
	}

	type esResp struct {
		Source portmodel.PortData `json:"_source"`
	}

	for key, port := range portsFromFile {
		resp, err := http.Get(elasticURL + "/" + testPortIndex + "/_doc/" + key)
		if err != nil {
			t.Errorf("could not get doc [%v]: %v", key, err)
			t.Fail()
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("could not read body [%v]: %v", key, err)
			t.Fail()
		}

		gotPort := &esResp{}
		if err := json.Unmarshal(body, gotPort); err != nil {
			t.Errorf("could not read body [%v]: %v", key, err)
			t.Fail()
		}

		if !reflect.DeepEqual(port, gotPort.Source) {
			t.Errorf("got %v expected %v", gotPort, port)
			t.Fail()
		}
	}
}
