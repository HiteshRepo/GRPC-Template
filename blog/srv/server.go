package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HiteshRepo/grpc-go-course/blog/blogpb"
	"github.com/HiteshRepo/grpc-go-course/blog/constants"
	"github.com/HiteshRepo/grpc-go-course/blog/logger"
	"github.com/HiteshRepo/grpc-go-course/blog/middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var collection *mongo.Collection

type server struct {
	collection *mongo.Collection
}

type blogItem struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string `bson:"author_id"`
	Title string `bson:"title"`
	Content string `bson:"content"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Println("Create blog request")
	blog := req.GetBlog()

	data := blogItem{AuthorID: blog.GetAuthorId(), Title: blog.GetTitle(), Content: blog.GetContent()}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error: %v", err),
			)
	}
	if oid,ok := res.InsertedID.(primitive.ObjectID); ok {
		return &blogpb.CreateBlogResponse{Blog: &blogpb.Blog{
			Id: oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title: blog.GetTitle(),
			Content: blog.GetContent(),
		}}, nil
	}
	return nil, status.Errorf(
		codes.Internal,
		fmt.Sprintf("Unable to extract inserted id: %v", err),
	)
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Println("Read blog request")
	blogId := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse blog-id: %v", err),
		)
	}

	data := blogItem{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res := collection.FindOne(ctx, bson.M{"_id": oid})
	if err = res.Decode(&data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find with given blog-id, err: %v", err),
		)
	}

	blogData := blogpb.Blog{}
	b,_ := json.Marshal(data)
	json.Unmarshal(b, &blogData)
	return &blogpb.ReadBlogResponse{Blog: &blogData}, nil


	//cursor, err := collection.Find(ctx, bson.D{})
	//if err != nil {
	//	return nil, status.Errorf(
	//		codes.Internal,
	//		fmt.Sprintf("Internal error: %v", err),
	//	)
	//}
	//if err != nil {
	//	defer cursor.Close(ctx)
	//	return nil, status.Errorf(
	//		codes.Internal,
	//		fmt.Sprintf("Internal error: %v", err),
	//	)
	//
	//} else {
	//	// iterate over docs using Next()
	//	for cursor.Next(ctx) {
	//		// declare a result BSON object
	//		var result bson.M
	//		err := cursor.Decode(&result)
	//
	//		// If there is a cursor.Decode error
	//		if err != nil {
	//			log.Println("cursor.Next() error:", err)
	//			os.Exit(1)
	//
	//			// If there are no cursor.Decode errors
	//		} else {
	//			b,_ := bson.Marshal(result)
	//			bson.Unmarshal(b, &data)
	//			if data.ID == oid {
	//				break
	//			}
	//		}
	//	}
	//	blogData := blogpb.Blog{}
	//	b,_ := json.Marshal(data)
	//	json.Unmarshal(b, &blogData)
	//	return &blogpb.ReadBlogResponse{Blog: &blogData}, nil
	//}
}

func (s *server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Println("Update blog request")
	blog := req.GetBlog()

	data := blogItem{AuthorID: blog.GetAuthorId(), Title: blog.GetTitle(), Content: blog.GetContent()}

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse blog-id: %v", err),
		)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := collection.ReplaceOne(ctx, bson.M{"_id": oid}, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update blog, err: %v", err),
		)
	}

	if res.MatchedCount > 0 {
		if res.ModifiedCount > 0 {
			res, err := s.ReadBlog(ctx, &blogpb.ReadBlogRequest{BlogId: req.GetBlog().GetId()})
			if err != nil {
				return &blogpb.UpdateBlogResponse{Status: "Blog was modified."}, nil
			} else {
				return &blogpb.UpdateBlogResponse{Status: fmt.Sprintf("Modified blog details: %v", res.GetBlog())}, nil
			}
		} else {
			return &blogpb.UpdateBlogResponse{Status: fmt.Sprintf("no blog was modified for given id : %s", req.GetBlog().GetId())}, nil
		}
	} else {
		return &blogpb.UpdateBlogResponse{Status: fmt.Sprintf("no blog was found for given id : %s", req.GetBlog().GetId())}, nil
	}
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	log.Println("Delete blog request")
	blogId := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse blog-id: %v", err),
		)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find with given blog-id, err: %v", err),
		)
	}

	if res.DeletedCount > 0 {
		return &blogpb.DeleteBlogResponse{Status: "Blog was deleted."}, nil
	} else {
		return &blogpb.DeleteBlogResponse{Status: fmt.Sprintf("no blog was found for given id : %s", req.GetBlogId())}, nil
	}
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	log.Println("List blog request")

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	defer cursor.Close(context.Background())
	// iterate over docs using Next()
	for cursor.Next(context.Background()) {
		// declare a result BSON object
		data := &blogItem{}
		err := cursor.Decode(data)

		// If there is a cursor.Decode error
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)
		}
		blogData := &blogpb.Blog{}
		b,_ := json.Marshal(data)
		json.Unmarshal(b, blogData)
		stream.Send(&blogpb.ListBlogResponse{Blog: blogData})
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	return nil
}

func main(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("connecting to mongodb")
	mongoUri := constants.MONGO_URI
	if len(os.Getenv("MONGO_URI")) > 0 {
		mongoUri = os.Getenv("MONGO_URI")
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	defer func() {
		log.Println("closing mongodb connection")
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	dbName := constants.DB_NAME
	dbCollection := constants.DB_COLLECTION
	if len(os.Getenv("DB_NAME")) > 0 {
		dbName = os.Getenv("DB_NAME")
	}
	if len(os.Getenv("DB_COLLECTION")) > 0 {
		dbCollection = os.Getenv("DB_COLLECTION")
	}
	collection = client.Database(dbName).Collection(dbCollection)

	log.Println("Blog service is starting")

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
	blogpb.RegisterBlogServiceServer(s, &server{collection: collection})

	if err := logger.Init(-1, "2006-01-02T15:04:05Z07:00"); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(s)

	grpc_prometheus.Register(s)
	middleware.RunPrometheusServer()

	go func() {
		log.Println("starting listener")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v",err)
		}
	}()

	// wait for ctrl+c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch

	log.Println("Stopping the server")
	s.Stop()
	log.Println("Closing the listener")
	lis.Close()
	log.Println("End of program")

}
