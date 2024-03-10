package store

import "context"

type Store interface {
	Connect(ctx context.Context) (bool, error)
	// Store persists the key value pair to the store. It ensures that it doesn't already exist; if it already does, it aborts.
	Store(ctx context.Context, id string, token any) error
	Retrieve(ctx context.Context, id string) (string, error)
	RetrieveAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, id string) (bool, error)
	Patch(ctx context.Context, id string, token any) (bool, error)
	Flush(ctx context.Context) (bool, error)
	Close(ctx context.Context) error
}
