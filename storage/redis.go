package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/shoksin/tracing-jaeger/models"
)

type NotesStorage struct {
	client redis.UniversalClient
}

func NewNoteStorage(client redis.UniversalClient) NotesStorage {
	return NotesStorage{client: client}
}

func (s NotesStorage) Store(ctx context.Context, note models.Note) error {
	data, err := json.Marshal(note)
	if err != nil {
		log.Println("storage(Store) Marhal:", err)
		return fmt.Errorf("marshal note: %w", err)
	}

	if err = s.client.Set(note.NoteID.String(), data, -1).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (s NotesStorage) Get(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	data, err := s.client.Get(noteID.String()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var note models.Note

	if err := json.Unmarshal(data, &note); err != nil {
		return nil, fmt.Errorf("unmarshal note: %w", err)
	}

	return &note, nil
}
