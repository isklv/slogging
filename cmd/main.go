package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isklv/slogging"
	http2 "github.com/isklv/slogging/http"
	ginmw "github.com/isklv/slogging/http/gin"
)

func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

	traceMW := ginmw.TraceMiddleware(log.Logger)

	//amqpTraceMW := amqpmw.TraceMiddleware(log.Logger)

	r := gin.New()
	r.Use(traceMW)

	r.GET("/hello", ginHelloWorld)

	r.Run(":8080")
}

type TestStruct struct {
	A string
	B int
	C bool
}

func ginHelloWorld(c *gin.Context) {
	ctx := c.Request.Context()
	slogging.L(ctx).Info("hello world")
	slogging.L(ctx).Error("bye bye world")

	slog.Info("so good")
	slogging.L(ctx).Fatal("fail")
	slogging.L(context.Background()).Error("empty context test")

	//slogging.L(ctx).Fatal("fatal test = )")
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	log := slogging.L(r.Context())
	log.Info("Тест HTTP ручки, тут должен быть TRACE заголовок")

	req, err := http.NewRequest("POST", "google.com", nil)
	if err != nil {
		slog.Info("message",
			slogging.ErrAttr(err))
	}

	req = http2.TraceRequest(r.Context(), req)
	log.Info("headers",
		slogging.StringAttr("xb-3trace", req.Header.Get("X-B3-TraceId")))

	w.Write([]byte("Hello world"))
}
