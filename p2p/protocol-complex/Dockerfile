FROM golang:1.10-alpine

WORKDIR /home/bzz

# ENV GOPATH /home/bzz

COPY . /home/bzz

RUN	apk add --update git bash gcc musl-dev linux-headers

RUN 	mkdir -p $GOPATH/src/github.com/ethereum && \
	cd $GOPATH/src/github.com/ethereum && \
	git clone https://github.com/nolash/go-ethereum && \
	cd go-ethereum && \
	git checkout sos18-demo-resource && \
	cd /home/bzz && \
	go build -o main main.go

CMD [ "bash" ]
