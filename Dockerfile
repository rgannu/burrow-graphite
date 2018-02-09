FROM golang

COPY . /go/src/github.com/rgannu/burrow-graphite
WORKDIR /go/src/github.com/rgannu/burrow-graphite

RUN go get ./
RUN go build

ENV BURROW_ADDR http://burrow-host:8000
ENV GRAPHITE_HOST graphite
ENV GRAPHITE_PORT 2003
ENV INTERVAL 30
CMD ./burrow-graphite --burrow-addr $BURROW_ADDR --graphite-host $GRAPHITE_HOST --graphite-port $GRAPHITE_PORT --interval $INTERVAL
