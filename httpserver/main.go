package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

const TimeOut = 1 * time.Second

func getRedisClient() *redis.Client {
	addr, ok := os.LookupEnv("REDIS")
	if !ok {
		addr = "192.168.1.8:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // use default Addr
		Password: "",   // no password set
		DB:       0,    // use default DB
	})

	return rdb
}

type RedisResponse struct {
	Payload string
	Error   error
}

func requestOnce(rdb *redis.Client, reqMessage redis.Message) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		panic("UUID failed")
	}

	ret := make(chan RedisResponse)

	channel := reqMessage.Channel
	reqMessage.Channel = channel + ":" + id.String()

	go func() {
		pubsub := rdb.Subscribe(reqMessage.Channel)
		defer pubsub.Close()

		for {
			// ReceiveTimeout is a low level API. Use ReceiveMessage instead.
			msgi, err := pubsub.ReceiveTimeout(TimeOut)
			if err != nil {
				ret <- RedisResponse{
					Error: err,
				}
				return
			}

			switch msg := msgi.(type) {
			case *redis.Subscription:
				fmt.Println("subscribed to", reqMessage.Channel)

				_, err := rdb.Publish(channel, reqMessage.String()).Result()
				if err != nil {
					ret <- RedisResponse{
						Error: err,
					}
					return
				}
			case *redis.Message:
				fmt.Println("received", msg.Payload, "from", msg.Channel)
				ret <- RedisResponse{
					Payload: msg.Payload,
				}
				return
			default:
				panic("unreached")
			}
		}
	}()

	data := <-ret
	return data.Payload, data.Error
}

func main() {
	router := gin.Default()

	rdb := getRedisClient()
	pong, err := rdb.Ping().Result()
	fmt.Println(pong, err)

	router.POST("/:service", func(c *gin.Context) {
		service := c.Param("service")

		body, err := c.GetRawData()

		if err != nil {
			c.JSON(http.StatusBadRequest, "Bad Request")
			return
		}

		payload, err := json.Marshal(gin.H{
			"Content-Type": c.GetHeader("Content-Type"),
			"Body":         body,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, "Bad Request")
			return
		}

		data, err := requestOnce(rdb, redis.Message{
			Channel: service,
			Payload: string(payload),
		})

		if err == nil {
			c.JSON(http.StatusOK, []byte(data))
		} else {
			fmt.Println(err.Error())
			c.JSON(http.StatusServiceUnavailable, []byte("Unreachable Service"))
		}
	})

	router.GET("/", func(c *gin.Context) {
		contentType := "text/html"
		host := c.Request.Host

		data, err := requestOnce(rdb, redis.Message{
			Channel: "html:" + host,
			Payload: "",
		})

		if err == nil {
			println(data)
			c.Data(http.StatusOK, contentType, []byte(data))
		} else {
			fmt.Println(err.Error())
			c.Data(http.StatusServiceUnavailable, contentType, []byte("Unreachable Service"))
		}
	})

	// Listen and serve on 0.0.0.0:8080
	router.Run(":80")
}
