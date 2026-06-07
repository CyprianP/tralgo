package main

import (
	"context"
	"log"
	"tralgo/config"
	"tralgo/router_mapping"

	"github.com/jackc/pgx/v5/pgxpool"
	"goyave.dev/goyave/v5"
)

func main() {

	pool, err := pgxpool.New(context.Background(), config.DatabaseUrl)
	if err != nil {
		log.Fatal("db connect error:", err)
	}
	defer pool.Close() // close connection when main finishes

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("db ping error:", err)
	}

	log.Println("connected to database")

	opts := goyave.Options{}
	server, err := goyave.New(opts)
	if err != nil {
		log.Fatal("server init error:", err)
	}

	server.RegisterRoutes(router_mapping.CreateRoutes(pool))
	server.Logger.Info("Registering hooks")
	server.RegisterSignalHook()
	server.RegisterStartupHook(func(s *goyave.Server) {
		s.Logger.Info("Server is listening", "host", s.Host())
	})

	server.RegisterShutdownHook(func(s *goyave.Server) {
		s.Logger.Info("Server is shutting down")
	})

	if err := server.Start(); err != nil {
		log.Fatal("server start error:", err)
	}

}
