FROM golang:1.5.1

RUN mkdir -p /go/src/github.com/ovh/tat
WORKDIR /go/src/github.com/ovh/tat

# this will ideally be built by the ONBUILD below ;)
CMD ["go-wrapper", "run"]

COPY . /go/src/github.com/ovh/tat
RUN go-wrapper download
RUN go-wrapper install
