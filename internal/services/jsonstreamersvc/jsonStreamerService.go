package jsonstreamersvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	ErrNotReadable      = errors.New("directory is not readable")
	ErrCouldNotOpenFile = errors.New("could not open")
	ErrDecodeDelimiter  = errors.New("could not find expected delimiter")
	ErrDecodeItem       = errors.New("could not decode item")
)

type JSONStreamer[T any] struct {
	fileToScan string
	stream     chan *Entry[T]
	stop       bool
}

type Entry[T any] struct {
	Data T
	Key  string
	Err  error
}

func New[T any](
	fileToScan string,
	channelBufferCapacity int,
) (*JSONStreamer[T], error) {
	if _, err := os.Stat(fileToScan); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotReadable, err)
	}

	return &JSONStreamer[T]{
		fileToScan: fileToScan,
		stream:     make(chan *Entry[T], channelBufferCapacity),
		stop:       false,
	}, nil
}

func (js *JSONStreamer[T]) Watch() <-chan *Entry[T] {
	return js.stream
}

func (js *JSONStreamer[T]) Start() {
	defer close(js.stream)

	f, err := os.Open(js.fileToScan)
	if err != nil {
		js.stream <- &Entry[T]{Err: fmt.Errorf("%w: [%s] %v", ErrCouldNotOpenFile, js.fileToScan, err)}
		return
	}

	defer f.Close()

	decoder := json.NewDecoder(f)
	if _, err = decoder.Token(); err != nil {
		js.stream <- &Entry[T]{Err: fmt.Errorf("%w: [opening] %v", ErrDecodeDelimiter, err)}
		return
	}

	for decoder.More() {
		if !shouldContinue() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if js.shouldExit() {
			break
		}

		// At first let's parse the key since this is not a valid json anymore
		keyB, tokenErr := decoder.Token()
		if tokenErr != nil {
			js.stream <- &Entry[T]{Err: fmt.Errorf("%w: %v", ErrDecodeItem, err)}
			return
		}

		key, ok := keyB.(string)
		if !ok {
			js.stream <- &Entry[T]{Err: fmt.Errorf("%w: could not convert key [%v] to string", ErrDecodeItem, keyB)}
		}

		data := new(T)
		if err = decoder.Decode(data); err != nil {
			js.stream <- &Entry[T]{Err: fmt.Errorf("%w: %v", ErrDecodeItem, err)}
			return
		}

		js.stream <- &Entry[T]{Data: *data, Key: key}
	}

	if _, err = decoder.Token(); err != nil {
		js.stream <- &Entry[T]{Err: fmt.Errorf("%w: [closing] %v", ErrDecodeDelimiter, err)}
	}
}

func shouldContinue() bool {
	// todo:check current ram consumption
	return true
}

func (js *JSONStreamer[T]) shouldExit() bool {
	// todo: implement graceful shutdown
	return false
}
