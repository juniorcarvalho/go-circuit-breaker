package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"net/http"
	"time"
)

var startTime time.Time = time.Now()

func server() {
	e := gin.Default()
	e.GET("/ping", func(ctx *gin.Context) {
		if time.Since(startTime) < 5*time.Second {
			ctx.String(http.StatusInternalServerError, "pong")
			return
		}
		ctx.String(http.StatusOK, "pong")
	})

	fmt.Printf("Stating server at port 8000\n")
	err := e.Run(":8000")
	if err != nil {
		return
	}
}

func DoReq() error {
	resp, err := http.Get("http://localhost:8000/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("bad response")
	}
	return nil
}

func main() {
	go server()

	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "test-cb",
			MaxRequests: 3,
			Timeout:     3 * time.Second,
			Interval:    1 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures == 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				fmt.Printf("CircuitBreaker %s change from %s to %s\n", name, from, to)
			},
			IsSuccessful: func(err error) bool {
				if err != nil {
					fmt.Printf("IsSuccessful: %s\n", err.Error())
				}
				return err == nil
			},
		},
	)

	fmt.Println("Call with circuit breaker")

	for i := 0; i < 100; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			err := DoReq()
			return nil, err
		})
		if err != nil {
			fmt.Printf("msg: %s\n", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
