package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
)

type server struct {

}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	log.Printf("Received Sum RPC: %v\n", req)
	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNumber()

	result := firstNumber + secondNumber
	return &calculatorpb.SumResponse{
		Result: result,
	}, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	log.Printf("Received PrimeNumberDecomposition RPC: %v\n", req)
	number := req.GetNumber()

	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number = number/divisor
		} else {
			divisor++
		}
	}
	return nil
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	log.Println("FindMaximum function was invoked with a streaming request...")
	max := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		number := req.GetNumber()
		if number > max {
			max = number
		}
		err = stream.Send(&calculatorpb.FindMaximumResponse{Maximum: max})
		if err != nil {
			log.Fatalf("Error while sending to client: %v", err)
			return err
		}
	}
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	log.Println("Received ComputeAverage RPC")

	sum := 0
	count := 0

	for {
		req,err := stream.Recv()
		if err == io.EOF {
			average := float64(sum)/float64(count)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{Average: average})
		}
		if err != nil {
			log.Fatalf("error while reading ComputeAverage client stream: %v", err)
		}
		sum += int(req.GetNumber())
		count += 1
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	log.Printf("Received SquareRoot RPC: %v\n", req)
	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Recieved a negative number: %v", number))
	}

	result := math.Sqrt(float64(number))
	return &calculatorpb.SquareRootResponse{
		SquareRoot: result,
	}, nil
}

func main(){
	lis,err := net.Listen("tcp", "0.0.0.0:50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v",err)
	}
}