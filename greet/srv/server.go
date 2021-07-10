package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/greet/constants"
	"github.com/HiteshRepo/grpc-go-course/greet/greetpb"
	"github.com/HiteshRepo/grpc-go-course/greet/logger"
	"github.com/HiteshRepo/grpc-go-course/greet/middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type server struct {

}

func (*server) Greet(ctx context.Context, req *greetpb.GreetingRequest) (*greetpb.GreetingResponse, error) {
	log.Printf("Greet function was invoked with req: %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()

	result := fmt.Sprintf("Hello, %s %s", firstName, lastName)
	return &greetpb.GreetingResponse{
		Result: result,
	}, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes function was invoked with req: %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	for i:=0; i<10; i++ {
		result := fmt.Sprintf("Hello, %s %s %s", firstName, lastName, strconv.Itoa(i))
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	log.Println("LongGreet function was invoked with a streaming request...")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()
		result += fmt.Sprintf("Hello, %s %s ", firstName, lastName)
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("GreetEveryone function was invoked with a streaming request...")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()
		result := fmt.Sprintf("Hello, %s %s ", firstName, lastName)
		time.Sleep(2000 * time.Millisecond)
		err = stream.Send(&greetpb.GreetEveryoneResponse{Result: result})
		if err != nil {
			log.Fatalf("Error while sending to client: %v", err)
			return err
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetingWithDeadlineRequest) (*greetpb.GreetingWithDeadlineResponse, error) {
	log.Printf("GreetWithDeadline function was invoked with req: %v\n", req)
	for i:=0; i<3; i++ {
		if ctx.Err() == context.Canceled {
			log.Println("Client canceled the request")
			return nil, status.Errorf(codes.Canceled, "Client canceled the request")
		}
		time.Sleep(1*time.Second)
	}
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()

	result := fmt.Sprintf("Hello, %s %s", firstName, lastName)
	return &greetpb.GreetingWithDeadlineResponse{
		Result: result,
	}, nil
}

func main(){
	fmt.Println("Starting greet server")

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
	greetpb.RegisterGreetServiceServer(s, &server{})

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
