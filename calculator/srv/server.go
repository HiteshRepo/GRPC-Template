package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/calculator/calculatorpb"
	"github.com/HiteshRepo/grpc-go-course/calculator/constants"
	"github.com/HiteshRepo/grpc-go-course/calculator/logger"
	"github.com/HiteshRepo/grpc-go-course/calculator/middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
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
	log.Println("Starting calculator server")

	servAddr := constants.GRPC_SRV_ADDR
	if len(os.Getenv("GRPC_SRV_ADDR")) > 0 {
		servAddr = os.Getenv("GRPC_SRV_ADDR")
	}
	lis,err := net.Listen("tcp", servAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// tls is false, because of an error in client side :
	// connection error: desc = "transport: authentication handshake failed: x509: certificate relies
	// on legacy Common Name field, use SANs or temporarily enable Common Name matching with GODEBUG=x509ignoreCN=0"
	opts := []grpc.ServerOption{}
	tlsOrNot := constants.TLS
	if len(os.Getenv("TLS")) > 0 {
		tlsOrNot = os.Getenv("TLS")
	}
	tls, _ := strconv.ParseBool(tlsOrNot)
	if tls {
		certFilePath := constants.SSL_CERT_PATH
		keyFilePath := constants.SSL_KEY_PATH
		if len(os.Getenv("SSL_CERT_PATH")) > 0 {
			certFilePath = os.Getenv("SSL_CERT_PATH")
		}
		if len(os.Getenv("SSL_KEY_PATH")) > 0 {
			keyFilePath = os.Getenv("SSL_KEY_PATH")
		}
		certFile := certFilePath
		keyFile := keyFilePath
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed to load certificates: %v", sslErr)
			return
		}

		opts = append(opts, grpc.Creds(creds))
	}
	opts = append(opts, middleware.GetGrpcMiddlewareOpts()...)
	s := grpc.NewServer(opts...)
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := logger.Init(-1, "2006-01-02T15:04:05Z07:00"); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(s)

	grpc_prometheus.Register(s)
	middleware.RunPrometheusServer()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v",err)
	}
}
