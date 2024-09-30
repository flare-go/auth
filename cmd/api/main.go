package main

import (
	"context"
	"log"
	"net"

	pb "goflare.io/auth/proto/pb"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedServiceServer
}

func (s *grpcServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Results: "Hello Shop"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServiceServer(s, &grpcServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
