package main

import (
	"context"
	"google.golang.org/grpc"
	"grpc-demo/client/pool"
	"grpc-demo/proto"
	"log"
	"sync"
	"time"
)

func main() {
	concurrentNum := 10000
	log.Printf("测试 %d 并发请求 ...",concurrentNum)
	Run1(concurrentNum)
	Run2(concurrentNum)
}

func Run1(concurrentNum int) {
	conn, err := grpc.Dial(":8888", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := proto.NewCalcSvcClient(conn)
	wg := sync.WaitGroup{}
	begin := time.Now()
	wg.Add(concurrentNum)
	for i := 0; i < concurrentNum; i++ {
		go func() {
			_, err := client.Sum(context.Background(), &proto.SumReq{
				A: 5,
				B: 10,
			})
			if err != nil {
				log.Println(err)
			}
			// log.Printf("5 + 10 = %d", resp.GetSum())
			wg.Done()
		}()

	}
	wg.Wait()
	log.Printf("http2多路复用，用时：%v", time.Now().Sub(begin))
}

func Run2(concurrentNum int) {
	opts := pool.Options{
		MaxIdle:     100,
		MaxConn:     200,
		WaitTimeout: -1,
		Dial: func() (*grpc.ClientConn, error) {
			return grpc.Dial(":8888", grpc.WithInsecure())
		},
	}
	connPool := pool.NewPool(opts)

	wg := sync.WaitGroup{}
	begin := time.Now()

	wg.Add(concurrentNum)
	for i := 0; i < concurrentNum; i++ {
		go func() {

			conn := connPool.Get()
			if conn == nil {
				panic("nil conn")
			}
			defer connPool.Put(conn)
			client := proto.NewCalcSvcClient(conn)
			_, err := client.Sum(context.Background(), &proto.SumReq{
				A: 5,
				B: 10,
			})
			if err != nil {
				log.Println(err)
			}
			// log.Printf("5 + 10 = %d", resp.GetSum())
			wg.Done()
		}()

	}
	wg.Wait()
	log.Printf("使用连接池，用时：%v", time.Now().Sub(begin))

}
