package server

import (
	"errors"
	"github.com/shoksin/tracing-jaeger/ internal/models"
	"github.com/shoksin/tracing-jaeger/ internal/storage"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type FiberHandler struct {
	notesStorage storage.NotesStorage
	tracer       oteltrace.Tracer
}

func NewFiberHandler(notesStorage storage.NotesStorage, tracer oteltrace.Tracer) FiberHandler {
	return FiberHandler{notesStorage: notesStorage, tracer: tracer}
}

func (h FiberHandler) CreateNote(fiberctx *fiber.Ctx) error {
	ctx, span := h.tracer.Start(fiberctx.UserContext(), "CreateNote")
	defer span.End()

	input := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{}

	span.AddEvent("body parsinng")
	if err := fiberctx.BodyParser(&input); err != nil {
		return err
	}

	span.AddEvent("call notesStrorage.Store")
	noteID := uuid.New()
	err := h.notesStorage.Store(ctx, models.Note{
		NoteID:    noteID,
		Title:     input.Title,
		Content:   input.Content,
		CreatedAt: time.Now(),
	})
	if err != nil {
		span.RecordError(err, oteltrace.WithAttributes(attribute.String("Error Info", "Error while ttying to store note in DB")))
		span.SetStatus(codes.Error, err.Error())

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	span.AddEvent("write JSON")
	return fiberctx.JSON(map[string]any{
		"note_id": noteID,
	})
}

func (h FiberHandler) GetNote(fiberctx *fiber.Ctx) error {
	ctx, span := h.tracer.Start(fiberctx.UserContext(), "GetNote")
	defer span.End()

	span.AddEvent("parse note_id")
	noteID, err := uuid.Parse(fiberctx.Query("note_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	span.AddEvent("call redis storage")
	note, err := h.notesStorage.Get(ctx, noteID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	span.AddEvent("write note in JSON")
	return fiberctx.JSON(note)
}
