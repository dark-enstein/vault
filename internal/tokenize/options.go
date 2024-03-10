package tokenize

import "github.com/dark-enstein/vault/pkg/store"

type Options func(*Manager)

func WithStore(store store.Store) func(*Manager) {
	return func(manager *Manager) {
		manager.store = store
	}
}
