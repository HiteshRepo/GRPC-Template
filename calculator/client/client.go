package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strconv"
	"time"
)

func main()  {
	fmt.Println("Hello I'm client")
	cc, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()
	c := calculatorpb.NewCalculatorServiceClient(cc)
	//doSumUnary(c)
	doSqrtUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBidiStreaming(c)
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