package main

import (
	"log"

	"github.com/shoksin/tracing-jaeger/trace"

	"github.com/go-redis/redis"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/shoksin/tracing-jaeger/server"
	"github.com/shoksin/tracing-jaeger/storage"
)

func main() {
	app := fiber.New()
	app.Use(otelfiber.Middleware())

	_, err := trace.InitTracer("localhost:4318", "Note Service")
	if err != nil {
		log.Fatal("init tracer: ", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatal("create redis client: ", err)
	}

	handler := server.NewFiberHandler(storage.NewNoteStorage(client))

	app.Post("/create", handler.CreateNote)
	app.Get("/get", handler.GetNote)

	log.Fatal(app.Listen(":8080"))
}
