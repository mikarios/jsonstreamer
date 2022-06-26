FROM golang:1.18.3
ARG version
# Create and use a directory where our project will be build
WORKDIR /go/src/github.com/mikarios/jsonstreamer/
COPY . /go/src/github.com/mikarios/jsonstreamer/

RUN go mod vendor

#RUN GO111MODULE=on GOFLAGS=-mod=vendor CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o covid ./cmd/server
RUN go build -o /jsonParser ./cmd/

CMD [ "/jsonParser" ]