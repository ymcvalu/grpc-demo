package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"grpc-demo/proto"
	"log"
	"net"
	"time"
)

var _ proto.CalcSvcServer = new(CalcSvc)

type CalcSvc struct{}

func (CalcSvc) Sum(ctx context.Context, req *proto.SumReq) (resp *proto.SumResp, err error) {
	a := req.GetA()
	b := req.GetB()
	log.Println("request coming ...")

	time.Sleep(time.Second)
	return &proto.SumResp{
		Sum: a + b,
	}, err
}

func main() {
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	proto.RegisterCalcSvcServer(s, &CalcSvc{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
