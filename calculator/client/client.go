package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/calculator/calculatorpb"
	"github.com/HiteshRepo/grpc-go-course/calculator/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func main()  {
	fmt.Println("Starting Calculator client")

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
	c := calculatorpb.NewCalculatorServiceClient(cc)
	//doSumUnary(c)
	//doSqrtUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBidiStreaming(c)
}

func doSumUnary(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.SumRequest {
		FirstNumber: 2,
		SecondNumber: 3,
	}
	res,err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Sum rpc : %v", err)
	}
	log.Printf("reponse from Sum rpc: %v", res.GetResult())
}

func doSqrtUnary(c calculatorpb.CalculatorServiceClient) {
	reqs := []*calculatorpb.SquareRootRequest {
		{Number: 16},
		{Number: -25},
	}
	for _,r := range reqs {
		req := r
		res,err := c.SquareRoot(context.Background(), req)
		if err != nil {
			respErr, ok := status.FromError(err)
			if ok {
				log.Println(respErr.Code().String() + " : " +respErr.Message())
			} else {
				log.Fatalf("error while calling Sum rpc : %v", err)
			}
		} else {
			log.Printf("reponse from SquareRoot rpc: %v", res.GetSquareRoot())
		}
	}
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberDecompositionRequest {
		Number: 123908765321234511,
	}
	resStream,err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumberDecomposition rpc : %v", err)
	}
	for {
		msg,err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading PrimeNumberDecomposition rpc : %v", err)
		}
		log.Printf("prime factor : %s", strconv.Itoa(int(msg.GetPrimeFactor())))
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	stream,err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error while opening ComputeAverage stream : %v", err)
	}

	numbers := []int32{3,5,9,54,32}

	for _,number := range numbers{
		log.Printf("Sending number: %v\n", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{Number: number})
	}

	res,err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while receiving CompiteAverage rpc response: %v", err)
	}

	log.Printf("The average is: %v\n", res.GetAverage())
}

func doBidiStreaming(c calculatorpb.CalculatorServiceClient)  {
	log.Println("starting bidi streaming.....")

	numbers := []int32{3,51,93,54,32}

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("error while invoking FindMaximum rpc: %v", err)
	}

	waitChan := make(chan struct{})

	go func() {
		for _,num := range numbers {
			log.Printf("sending message: %v\n", num)
			stream.Send(&calculatorpb.FindMaximumRequest{Number: num})
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
				log.Fatalf("Error while receiving from FindMaximum rpc response: %v", err)
				break
			}
			log.Printf("response received: %v\n", res.GetMaximum())
		}
		close(waitChan)
	}()

	<-waitChan
}