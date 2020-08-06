package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v2"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)


type ratelimitService struct{}

func (r ratelimitService) ShouldRateLimit(ctx context.Context, request *pb.RateLimitRequest) (*pb.RateLimitResponse, error) {
	now := (time.Now().Unix()/60)*60
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()
	var uq string
	if request.Descriptors[0].Entries[0].Value!=""{
		uq = request.Domain+"_"+request.Descriptors[0].Entries[0].Key +"_" +request.Descriptors[0].Entries[0].Value
	}else {
		uq = request.Domain+"_"+request.Descriptors[0].Entries[0].Key
	}
	uq+=fmt.Sprint(now)
	reply,err:=redis.String(conn.Do("GET", uq))
	if  err!= nil&& reply!="" {
		return &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_UNKNOWN,}, err
	}
	if count,_:=strconv.Atoi(fmt.Sprint(reply));count>2 {
		return  &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_OVER_LIMIT}, nil
	}
	if _, err := conn.Do("INCR", uq); err != nil {
		return &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_UNKNOWN}, err
	}
	return  &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_OK,}, nil
}

func main() {
	log.SetFlags(log.Llongfile)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
	}
	s := grpc.NewServer()
	pb.RegisterRateLimitServiceServer(s, &ratelimitService{})
	reflection.Register(s)
	s.Serve(listener)
}
