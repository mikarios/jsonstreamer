package jsonstreamersvc

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
)

var testDataFile = "../../../third_party/portsSmall.json"

func TestNew(t *testing.T) {
	t.Parallel()

	type args struct {
		jsonToScan            string
		channelBufferCapacity int
		notifyOnFinish        chan<- interface{}
	}

	notifyOnFinish := make(chan interface{})

	tests := []struct {
		name    string
		args    args
		want    *JSONStreamer[*portmodel.PortData]
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				jsonToScan:            testDataFile,
				channelBufferCapacity: 0,
				notifyOnFinish:        notifyOnFinish,
			},
			want: &JSONStreamer[*portmodel.PortData]{
				fileToScan:     testDataFile,
				stop:           false,
				notifyOnFinish: notifyOnFinish,
			},
			wantErr: false,
		},
		{
			name: "invalid file",
			args: args{
				jsonToScan:            "invalidFile",
				channelBufferCapacity: 0,
				notifyOnFinish:        notifyOnFinish,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := New[*portmodel.PortData](tt.args.jsonToScan, tt.args.channelBufferCapacity, tt.args.notifyOnFinish)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// ignore generated stream for the tests
			if got != nil {
				got.stream = nil
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONStreamer_Watch(t *testing.T) {
	t.Parallel()

	type MyType string

	type fields struct {
		stream chan *Entry[MyType]
	}

	expectedChan := make(chan *Entry[MyType])

	tests := []struct {
		name   string
		fields fields
		want   <-chan *Entry[MyType]
	}{
		{
			name: "happy path",
			fields: fields{
				stream: expectedChan,
			},
			want: expectedChan,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			js := &JSONStreamer[MyType]{
				stream: tt.fields.stream,
			}
			if got := js.Watch(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Watch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONStreamer_Start(t *testing.T) {
	t.Parallel()

	portsFromStream := make(map[string]portmodel.PortData)
	finish := make(chan interface{})

	js, err := New[*portmodel.PortData](testDataFile, 0, finish)
	if err != nil {
		t.Error("error should be nil")
		t.Fail()
	}

	incoming := js.Watch()

	go js.Start()

	for entry := range incoming {
		portsFromStream[entry.Key] = *entry.Data
	}

	select {
	case <-time.After(1 * time.Second):
		t.Error("did not receive finish signal")
		t.Fail()
	case <-finish:
	}

	var portsFromFile map[string]portmodel.PortData

	f, _ := os.Open(testDataFile)

	if err = json.NewDecoder(f).Decode(&portsFromFile); err != nil {
		t.Error("could not decode file")
		t.Fail()
	}

	if !reflect.DeepEqual(portsFromFile, portsFromStream) {
		t.Errorf("not parsed correctly")
		t.Fail()
	}
}
