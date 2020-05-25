# 　GRPC 服务调用实践


##　contOS7， go 1.13

    yum install autoconf automake libtool gcc gcc-c++ zlib-devel
    
#### 1 安装　protobuf
   
     protoc --version 　# 能看到 版本信息就安装成功了

#### 2 安装 protobuf golang 插件

     go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
            
#### 3 目录列表

    tree
    .
    ├── go.mod
    ├── go.sum
    ├── pb
    │   └── hello.proto
    └── README.md


#### hello.proto　内容如下

    syntax = "proto3";
    
    package hello;
    
    // 定义服务
    service Hello {
      rpc SayHello (HelloRequest) returns (HelloReply) {}
    }
    
    // 请求体的结构体
    message HelloRequest {
    
      string name = 1;
    
    }
    
    // 响应的结构体
    message HelloReply {
      string message = 1;
      int64 code = 2;
    }

##### 在pb目录下运行下面命令会生成ＧＯ文件　hello.pb.go

    protoc --go_out=plugins=grpc:. hello.proto 
    
    2020/05/25 21:08:52 WARNING: Missing 'go_package' option in "hello.proto",
    please specify it with the full Go package path as
    a future release of protoc-gen-go will require this be specified.
    See https://developers.google.com/protocol-buffers/docs/reference/go-generated#package for more information.


#### 生成文件后的目录如下

    tree
    .
    ├── go.mod
    ├── go.sum
    ├── pb
    │   ├── hello.pb.go
    │   └── hello.proto
    └── README.md

#### 创建服务端go程序，代码如下

    mkdir service
    cd service
    
vi server.go
    
       
       package main
       
       import (
           "golang.org/x/net/context"
           "google.golang.org/grpc"
           "google.golang.org/grpc/reflection"
           "log"
           "net"
           pb "rpc-hello/pb"
       )
       
       // server 用来实现 hello.HelloServer
       
       type server struct{}
       
       // 实现 hello.SayHello 方法
       
       // (context.Context, *HelloRequest) (*HelloReply, error)
       
       func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
           return &pb.HelloReply{Message: "Hello " + in.Name, Code: 200}, nil
       }
       
       func main() {
       
           lis, err := net.Listen("tcp", ":60051")
       
           if err != nil {
               log.Fatalf("failed to listen: %v", err)
           }
       
           s := grpc.NewServer()
       
           pb.RegisterHelloServer(s, &server{})
       
           //在 server 中 注册 gRPC 的 reflection service
       
           reflection.Register(s)
       
           if err := s.Serve(lis); err != nil {
               log.Fatalf("failed to serve: %v", err)
           }
       }
       
#### 创建客户端go程序，代码如下
    
    mkdir client
    cd client

vi client.go

        package main
        
        import (
            "fmt"
            "github.com/gin-gonic/gin"
            "google.golang.org/grpc"
            "log"
            "net/http"
            pb "rpc-hello/pb"
        )
        
        func main() {
        
            r := gin.Default()
        
            r.GET("/rpc/hello", func(c *gin.Context) {
                sayHello(c)
            })
        
            // Run http server
            if err := r.Run(":8058"); err != nil {
                log.Fatalf("could not run server: %v", err)
            }
        }
        
        func sayHello(c *gin.Context) {
        
            // Set up a connection to the server.
            conn, err := grpc.Dial("localhost:60051", grpc.WithInsecure())
            if err != nil {
                log.Fatalf("did not connect: %v", err)
            }
        
            defer conn.Close()
        
            client := pb.NewHelloClient(conn)
            name := c.DefaultQuery("name","战士上战场")
            req := &pb.HelloRequest{Name: name}
            res, err := client.SayHello(c, req)
        
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": err.Error(),
                })
                return
            }
        
            c.JSON(http.StatusOK, gin.H{
                "result": fmt.Sprint(res.Message),
                "code": fmt.Sprint(res.Code),
            })
        
        }
        
#### 最终目录如下

        tree
        .
        ├── client
        │   └── client.go
        ├── go.mod
        ├── go.sum
        ├── pb
        │   ├── hello.pb.go
        │   └── hello.proto
        ├── README.md
        └── service
            └── service.go


#### 运行方法步骤如下

1、 进入服务端的目录，打开一个命令终端
     
     cd service
     go run service.go

２、进入客户端目录，打开另一个命令终端
    
    cd client
    go run client.go

３、浏览器中打开网址
    
    http://127.0.0.1:8058/rpc/hello
    
客户端会输出信息

    [GIN-debug] Listening and serving HTTP on :8058
    [GIN] 2020/05/25 - 22:01:11 | 200 |      5.9897ms |   192.168.1.199 | GET      "/rpc/hello"
    [GIN] 2020/05/25 - 22:01:13 | 200 |    3.403369ms |   192.168.1.199 | GET      "/rpc/hello"
    [GIN] 2020/05/25 - 22:01:13 | 200 |    1.859842ms |   192.168.1.199 | GET      "/rpc/hello"
    [GIN] 2020/05/25 - 22:01:21 | 200 |    1.781744ms |   192.168.1.199 | GET      "/rpc/hello"

浏览器输出

    {"code":"200","result":"Hello GO微服务ＧＲＰＣ来了！"}

服务端终止后浏览器会输出

        {"error":"rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp [::1]:60051: connect: connection refused\""}
        
客户端输出500信息
        
        [GIN] 2020/05/25 - 22:04:15 | 200 |     4.28828ms |   192.168.1.199 | GET      "/rpc/hello"
        [GIN] 2020/05/25 - 22:04:20 | 500 |     919.957µs |   192.168.1.199 | GET      "/rpc/hello"
        [GIN] 2020/05/25 - 22:04:21 | 500 |     874.983µs |   192.168.1.199 | GET      "/rpc/hello"
        [GIN] 2020/05/25 - 22:04:21 | 500 |    1.429136ms |   192.168.1.199 | GET      "/rpc/hello"
        [GIN] 2020/05/25 - 22:04:26 | 500 |    6.232437ms |   192.168.1.199 | GET      "/rpc/hello"

重启服务端恢复正常
