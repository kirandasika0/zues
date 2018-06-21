FROM golang:1.10 as build-env
RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
WORKDIR $GOPATH/src/zues
COPY . .
RUN go get github.com/tools/godep
RUN CGO_ENABLED=0 go install -a std
RUN CGO_ENABLED=0 godep go build -ldflags '-d -w -s'
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
WORKDIR /zues
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build-env $GOPATH/src/zues/main .
EXPOSE 8284
ENTRYPOINT [ "./main" ]
