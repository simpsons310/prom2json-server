package main

import (
	"context"
	"net/http"
	"os/signal"
	p2jsvr "simpsons310/prom2json-server/internal"
	"syscall"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		done()
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	// load args
	args, err := p2jsvr.ParseArgument()
	if err != nil {
		panic(err)
	}

	// parse config
	cfg, err := p2jsvr.LoadConfig(args.ConfigFile)
	if err != nil {
		panic(err)
	}

	// new app
	app, err := p2jsvr.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// add logger context
	ctx = app.ContextWithLogger(ctx)

	// add http handler to server
	mux := http.NewServeMux()
	app.RegisterHandler(mux)

	// start server
	svr := p2jsvr.NewServer(cfg.Server)
	if err := svr.Start(ctx, mux); err != nil {
		panic(err)
	}
}
