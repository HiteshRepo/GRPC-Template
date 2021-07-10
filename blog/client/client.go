package main

import (
	"context"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/blog/blogpb"
	"github.com/HiteshRepo/grpc-go-course/blog/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
	"strconv"
)

func main()  {
	fmt.Println("Starting blog client")

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
	c := blogpb.NewBlogServiceClient(cc)
	//createBlog(c)
	//readBlog(c)
	//updateBlog(c)
	//deleteBlog(c)
	listBlog(c)
}

func createBlog(c blogpb.BlogServiceClient) {
	log.Println("starting create blog.....")

	blogs := []*blogpb.CreateBlogRequest{
		{
			Blog: &blogpb.Blog {
				Title:    "Auth using GRPC",
				Content:  "How to implement auth microservice in GRPC.",
				AuthorId: "717010",
			},
		},
		{
			Blog: &blogpb.Blog {
				Title:    "Rest vs GRPC",
				Content:  "Speaks about advantages of GRPC over REST.",
				AuthorId: "717010",
			},
		},
		{
			Blog: &blogpb.Blog {
				Title:    "K8 Microservices implementation",
				Content:  "An use case to demonstrate microservice setup using K8",
				AuthorId: "717010",
			},
		},
	}

	for _,req := range blogs {
		res,err := c.CreateBlog(context.Background(), req)
		if err != nil {
			log.Fatalf("error while calling create-blog rpc : %v", err)
		}
		log.Printf("blog created with details: %v", res.Blog)
	}
}

func readBlog(c blogpb.BlogServiceClient) {
	log.Println("starting read blog.....")
	req := &blogpb.ReadBlogRequest {
		BlogId: "60e9609f0cc000938ec8d025",
	}
	res,err := c.ReadBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling read-blog rpc : %v", err)
	}
	log.Printf("blog fetched with details: %v", res.Blog)
}

func updateBlog(c blogpb.BlogServiceClient) {
	log.Println("starting update blog.....")
	req := &blogpb.UpdateBlogRequest {
		Blog: &blogpb.Blog{
			Id: "60e9609f0cc000938ec8d025",
			Title: "Auth using GRPC",
			Content: "Content available....",
			AuthorId: "717010",
		},
	}
	res,err := c.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling update-blog rpc : %v", err)
	}
	log.Printf("blog update with status: %v", res.Status)
}

func deleteBlog(c blogpb.BlogServiceClient) {
	log.Println("starting delete blog.....")
	req := &blogpb.DeleteBlogRequest {
		BlogId: "60e9609f0cc000938ec8d025",
	}
	res,err := c.DeleteBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling delete-blog rpc : %v", err)
	}
	log.Printf("blog deleted with status: %v", res.Status)
}

func listBlog(c blogpb.BlogServiceClient) {
	log.Println("starting list blog streaming.....")
	stream,err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("error while calling list-blog stream rpc : %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Some error occurred: %v", err)
		}
		log.Println(res.GetBlog())
	}
}
