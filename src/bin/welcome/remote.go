package main

import (
	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"google.golang.org/grpc"
)

type LazyApiClient struct {
	client     pb.ApiServiceClient
	connection *grpc.ClientConn
	opts       []grpc.DialOption
	url        string
}

func (ac *LazyApiClient) Close() {
	if ac.connection != nil {
		ac.connection.Close()
	}
	ac.connection = nil
}

func (ac *LazyApiClient) Get() pb.ApiServiceClient {
	if ac.connection != nil {
		return ac.client
	}
	aconn, err := grpc.Dial(ac.url, ac.opts...)
	if err != nil {
		logrus.Fatalf("remote.go: did not connect to api service: %v", err)
	}
	ac.client = pb.NewApiServiceClient(aconn)
	ac.connection = aconn
	return ac.client
}

func NewLazyApiClient(url string, opts []grpc.DialOption) *LazyApiClient {
	return &LazyApiClient{
		opts: opts,
		url:  url,
	}
}
