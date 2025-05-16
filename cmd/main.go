package main

import (
	"context"
	"github.com/shoksin/tracing-jaeger/ internal/server"
	"github.com/shoksin/tracing-jaeger/ internal/storage"
	"github.com/shoksin/tracing-jaeger/ internal/trace"
	"log"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	app := fiber.New()
	app.Use(otelfiber.Middleware())

	tracer, err := trace.InitTracer("localhost:4318", "Note Service")
	if err != nil {
		log.Fatal("init tracer: ", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("create redis client: ", err)
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		log.Fatal("enable instrument tracing: ", err)
	}

	handler := server.NewFiberHandler(storage.NewNoteStorage(client), tracer)

	app.Post("/create", handler.CreateNote)
	app.Get("/get", handler.GetNote)

	log.Fatal(app.Listen(":8080"))
}
