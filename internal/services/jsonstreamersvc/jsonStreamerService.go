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
	ErrNotADirectory    = errors.New("path is not a directory")
	ErrCouldNotOpenFile = errors.New("could not open")
	ErrDecodeDelimiter  = errors.New("could not find expected delimiter")
	ErrDecodeItem       = errors.New("could not decode item")
	shutDownChannel     = make(chan interface{})
)

type JSONStreamer[T any] struct {
	fileToScan     string
	stream         chan *Entry[T]
	stop           bool
	notifyOnFinish chan<- interface{}
}

type Entry[T any] struct {
	Data T
	Key  string
	Err  error
}

func New[T any](
	jsonToScan string,
	channelBufferCapacity int,
	notifyOnFinish chan<- interface{},
) (*JSONStreamer[T], error) {
	if _, err := os.Stat(jsonToScan); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotReadable, err)
	}

	return &JSONStreamer[T]{
		fileToScan:     jsonToScan,
		stream:         make(chan *Entry[T], channelBufferCapacity),
		stop:           false,
		notifyOnFinish: notifyOnFinish,
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
			shutDownChannel <- struct{}{}
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
		if err := decoder.Decode(data); err != nil {
			js.stream <- &Entry[T]{Err: fmt.Errorf("%w: %v", ErrDecodeItem, err)}
			return
		}

		js.stream <- &Entry[T]{Data: *data, Key: key}
	}

	if _, err := decoder.Token(); err != nil {
		js.stream <- &Entry[T]{Err: fmt.Errorf("%w: [closing] %v", ErrDecodeDelimiter, err)}
	}

	go func() {
		js.notifyOnFinish <- struct{}{}
	}()
	go func() {
		shutDownChannel <- struct{}{}
	}()
}

func shouldContinue() bool {
	// todo:check current ram consumption
	return true
}

func (js *JSONStreamer[T]) shouldExit() bool {
	return js.stop
}

func (js *JSONStreamer[T]) GracefulShutdown() <-chan interface{} {
	js.stop = true
	return shutDownChannel
}
