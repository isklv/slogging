package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/isklv/slogging"
	chimw "github.com/isklv/slogging/http/chi"
)

func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

	r := chi.NewRouter()

	r.Use(chimw.TraceMiddleware(sl.Logger))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		slogging.L(r.Context()).Info("hello world")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	addr := ":8080"
	sl.Info("starting server", slogging.StringAttr("addr", addr))
	if err := http.ListenAndServe(addr, r); err != nil {
		sl.Error("server failed", slogging.ErrAttr(err))
	}
}
