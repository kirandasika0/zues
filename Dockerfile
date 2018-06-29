FROM golang:1.10 as build-env
RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
WORKDIR $GOPATH/src/zues
COPY . .
RUN CGO_ENABLED=0 go install -a std
RUN CGO_ENABLED=0 go build -ldflags '-d -w -s'
RUN CGO_ENABLED=0 GOOS=linux go build \ 
    -ldflags "-X main.version=$(git rev-parse HEAD)" \ 
    -a -installsuffix cgo -o main .


FROM alpine
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV DOCKER_ENV=true
ENV IN_CLUSTER=true
WORKDIR /zues
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir ~/.kube
COPY --from=build-env $GOPATH/src/zues/main .
COPY ./kubeconfig .
EXPOSE 8284
ENTRYPOINT [ "./main" ]
