# Zap Logger
Pkg Url : https://github.com/uber-go/zap
1. It is very fast in comparison to other golang logging packages.
2. It provides both structured logging and printf style logging.
3. It includes reflection-free, zero-allocation JSON encoder.
4. The base 'Logger' avoids serialization overhead and allocations wherever possible.
5. The base 'Logger' only supports strongly-typed, structured logging. Advisable to use where every microsecond and every allocation matter.
5. The 'SuggaredLogger' is a bit slower in comparison to base 'Logger' as it supports both structured and printf-style logging.

# Prometheus
Pkg Url : https://github.com/prometheus/client_golang
1. It is used as a monitoring tool.
2. It can provide insight on hardware and software metrics like response latency, resource overload, app errors, request count, request duration, etc.
3. Provides automated monitoring coupled with alerting.
4. Edge over other monitoring tools like NewRelic, datadog, etc. with polling strategy to pull metrics rather than pushing [by apps/services] to monitoring tools.
5. Human-readable text based format for displaying metrics.

# Realize
Pkg Url : https://github.com/oxequa/realize
1. This is used as a watcher like nodemon in nodejs.
2. High performance Live Reload.
3. Custom env variables for project.
4. Manage multiple projects at the same time.

# GRPC over Rest
1. Protobuf instead of JSON : Protobuf serializes and deserializes structured data to communicate via binary.
2. Built on HTTP 2 instead of HTTP 1.1 : supports bidirectional client-response communication.
3. Code generation features are native to gRPC via its in-built protoc compiler versus use a third-party tool such as Swagger to auto-generate the code for API calls.
4. gRPC is approx 7 times faster than REST when receiving data & almost 10 times faster than REST when sending data for a specific payload.

# grpc-template
### Pre-Requisite
- protoc must be installed - if you wish to generate code from .proto file
### Step 1 : Define communication protocol
- Define message types and services in a **<your-project>.proto** file. Example format : [proto-file](https://github.com/HiteshRepo/GRPC-Template/blob/master/blog/blogpb/blog.proto)
- Compile the **<your-project>.proto** file using a proto compiler **protoc** using command format : protoc --proto_path=<proto-file-path> --go_out=<output-dir>.
- After successful code-generation, a **<your-project>.pb.go** file will get created.

### Step 2 : Implementation of service interface
- In '.go' file where service needs to be implemented, import the auto-generated **<your-project>.pg.go** file. Please see : [service-implementation](https://github.com/HiteshRepo/GRPC-Template/blob/master/blog/srv/server.go).
- Implement <your-project>ServiceServer interface here.

### Step 3 : Create a client or use evans-cli
In order to test the services, either create a client or use evans-cli:
- Client implementation example : https://github.com/HiteshRepo/GRPC-Template/blob/master/blog/client/client.go
- evans-cli : https://github.com/ktr0731/evans