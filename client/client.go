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
	name := c.DefaultQuery("name","GO微服务ＧＲＰＣ来了！")
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
