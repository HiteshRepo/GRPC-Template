package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/greet/constants"
	"github.com/HiteshRepo/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func main()  {
	fmt.Println("Starting greet client")

	// tls is false, because of an error in client side :
	// connection error: desc = "transport: authentication handshake failed: x509: certificate relies
	// on legacy Common Name field, use SANs or temporarily enable Common Name matching with GODEBUG=x509ignoreCN=0"
	tlsOrNot := constants.TLS
	if len(os.Getenv("TLS")) > 0 {
		tlsOrNot = os.Getenv("TLS")
	}
	tls, _ := strconv.ParseBool(tlsOrNot)
	opts := grpc.WithInsecure()
	if tls {
		certFilePath := constants.SSL_CA_CERT_PATH
		if len(os.Getenv("SSL_CA_CERT_PATH")) > 0 {
			certFilePath = os.Getenv("SSL_CA_CERT_PATH")
		}
		certFile := certFilePath
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Failed to loading ca trust certificates: %v", sslErr)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	servAddr := constants.GRPC_SRV_ADDR
	if len(os.Getenv("GRPC_SRV_ADDR")) > 0 {
		servAddr = os.Getenv("GRPC_SRV_ADDR")
	}
	cc, err := grpc.Dial(servAddr, opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()
	c := greetpb.NewGreetServiceClient(cc)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBidiStreaming(c)
	//doUnaryWithDeadline(c, 5*time.Second) //should complete - server timeout 3s
	//doUnaryWithDeadline(c, 1*time.Second) //should timeout - server timeout 3s
}

func doUnary(c greetpb.GreetServiceClient) {
	log.Println("starting unary.....")
	req := &greetpb.GreetingRequest {
		Greeting: &greetpb.Greeting{
			FirstName: "Hitesh",
			LastName: "Pattanayak",
		},
	}
	res,err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling greet rpc : %v", err)
	}
	log.Printf("reponse from Greet rpc: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	log.Println("starting server streaming.....")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Hitesh",
			LastName: "Pattanayak",
		},
	}
	resStream,err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes rpc : %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream from GreetManyTimes rpc : %v", err)
		}
		log.Printf("Response form GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	log.Println("starting client streaming.....")
	requests := []*greetpb.LongGreetRequest{
		{Greeting: &greetpb.Greeting{FirstName: "Hitesh", LastName: "Pattanayak"}},
		{Greeting: &greetpb.Greeting{FirstName: "John", LastName: "Doe"}},
		{Greeting: &greetpb.Greeting{FirstName: "Sheldon", LastName: "Cooper"}},
		{Greeting: &greetpb.Greeting{FirstName: "Howard", LastName: "Wolowitz"}},
		{Greeting: &greetpb.Greeting{FirstName: "Leonard", LastName: "Hoffstador"}},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while invoking LongGreet rpc: %v", err)
	}

	for _,req := range requests {
		log.Printf("Sending request: %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while reading LongGreet response: %v", err)
	}
	log.Printf("LongGreet response: %v", res)
}

func doBidiStreaming(c greetpb.GreetServiceClient)  {
	log.Println("starting bidi streaming.....")

	requests := []*greetpb.GreetEveryoneRequest{
		{Greeting: &greetpb.Greeting{FirstName: "Hitesh", LastName: "Pattanayak"}},
		{Greeting: &greetpb.Greeting{FirstName: "John", LastName: "Doe"}},
		{Greeting: &greetpb.Greeting{FirstName: "Sheldon", LastName: "Cooper"}},
		{Greeting: &greetpb.Greeting{FirstName: "Howard", LastName: "Wolowitz"}},
		{Greeting: &greetpb.Greeting{FirstName: "Leonard", LastName: "Hoffstador"}},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error while invoking GreetEveryone rpc: %v", err)
	}

	waitChan := make(chan struct{})

	go func() {
		for _,req := range requests {
			log.Printf("sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving from GreetEveryone rpc response: %v", err)
				break
			}
			log.Printf("response received: %v\n", res.GetResult())
		}
		close(waitChan)
	}()

	<-waitChan
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	log.Println("starting doUnaryWithDeadline.....")
	req := &greetpb.GreetingWithDeadlineRequest {
		Greeting: &greetpb.Greeting{
			FirstName: "Hitesh",
			LastName: "Pattanayak",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res,err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println("Timeout exceeded")
			} else {
				log.Println(statusErr.Code().String() + " : " + statusErr.Message())
			}
		} else {
			log.Fatalf("error while calling GreetWithDeadline rpc : %v", err)
		}
	} else {
		log.Printf("reponse from GreetWithDeadline rpc: %v", res.Result)
	}
}