package store

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/dark-enstein/vault/internal/vlog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"sync"
)

// TODO: implement btrees for storage

type Gob struct {
	loc string
	// basin is a temporary store, since it will be written to many times by separate method. It should be treated the same.
	basin  *Map
	fd     *os.File
	logger *vlog.Logger
	sync.RWMutex
}

func NewGob(ctx context.Context, loc string, logger *vlog.Logger, trunc bool) (*Gob, error) {
	// switch trunc to using functonal options
	log := logger.Logger()

	// gob encode
	err := IsValidFile(loc, log)
	if err != nil {
		log.Error().Msgf("gob store destination: %s invalid: error: %s\n", loc, err.Error())
		return nil, err
	}

	// open file
	fd, err := os.OpenFile(loc, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Info().Msgf("error while creating file at location %s: %s\n", loc, err.Error())
		return nil, err
	}

	if trunc {
		// truncate the file to zero
		err := fd.Truncate(0)
		if err != nil {
			log.Info().Msgf("could not truncate existing gob store %s: %s\n", loc, err.Error())
			return nil, err
		}
	}
	return &Gob{loc, NewSyncMap(ctx, logger), fd, logger, sync.RWMutex{}}, nil
}

func (g *Gob) Connect(ctx context.Context) (bool, error) {
	return g.basin.Connect(ctx)
}

func (g *Gob) Store(ctx context.Context, id string, token any) error {
	log := g.logger.Logger()

	// TODO: Revisit this it's best to refresh before persisting new data; just in case there has been a latest update first refresh in-memory map, or the in-memory map has been cleared below
	// see below
	err := g.MapRefresh(ctx)
	if err != nil {
		log.Error().Msgf("error while refresh gob persistent storage: error: %s\n", err.Error())
		return err
	}

	// store new key value pair in the in-memory store
	// TODO: rename
	err = g.basin.Store(ctx, id, token)
	if err != nil {
		return err
	}

	// persist the in-memory store to disk
	i, err := g.persist(ctx, true)
	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("wrote 0 bytes to gob persistent storage")
	}

	// TODO: revisit this later: ?do we go for refreshing before every write, or ensuring not to clear the in-memory cache after successive writes?
	// I think its more idempotent and repeatable to clear the cache after every write, regardless of if it is manywrites or onewrite. A seperate implementation for writemany could factor this optimization in mind.
	// Potentially move this out from Store, into a pipeline function? So its order can be interspersed with additional logic. I don't think its atomic.
	b, err := g.basin.Flush(ctx)
	if !b || err != nil {
		return err
	}

	return nil
}

func (g *Gob) Patch(ctx context.Context, id string, token any) (bool, error) {
	log := g.logger.Logger()

	// first refresh in-memory map
	err := g.MapRefresh(ctx)
	if err != nil {
		log.Error().Msgf("error while refresh gob persistent storage: error: %s\n", err.Error())
		return false, err
	}

	// check if key exists
	b, err := g.basin.Patch(ctx, id, token)
	if err != nil || !b {
		log.Debug().Msgf("error with patching entry with id: %s\n", id)
		return false, fmt.Errorf("error with patching entry with id: %s\n", id)
	}

	// persist in-memory map
	i, err := g.persist(ctx, true)
	if err != nil {
		log.Debug().Msgf("error while persisting patched entry with id: %s\n", id)
		return false, fmt.Errorf("error while persisting patched entry with id: %s\n", id)
	}

	if i == 0 {
		return false, fmt.Errorf("wrote 0 bytes to gob persistent storage")
	}

	return true, nil
}

func (g *Gob) Retrieve(ctx context.Context, id string) (string, error) {
	log := g.logger.Logger()
	// first refresh in-memory map
	err := g.MapRefresh(ctx)
	if err != nil {
		log.Error().Msgf("error while refresh gob persistent storage: error: %s\n", err.Error())
		return "", err
	}

	// retrieve value if it exists in store map
	value, err := g.basin.Retrieve(ctx, id)
	if err != nil {
		log.Debug().Msgf("error while retrieving value with id: %s: %s\n", id, err.Error())
		return "", fmt.Errorf("error while retrieving value with id: %s: %s\n", id, err.Error())
	}

	return value, nil
}

// RetrieveAll
func (g *Gob) RetrieveAll(ctx context.Context) (map[string]string, error) {
	log := g.logger.Logger()

	// first refresh in-memory map
	err := g.MapRefresh(ctx)
	if err != nil {
		log.Error().Msgf("error while refresh gob persistent storage: error: %s\n", err.Error())
		return nil, err
	}

	// then retrieve in-memory: opportunities for optimization here
	m, err := g.basin.RetrieveAll(ctx)
	if err != nil {
		log.Debug().Msgf("error while retrieving all entries: %s\n", err.Error())
		return nil, fmt.Errorf("error while retrieving all entries: %s\n", err.Error())
	}

	return m, err
}

func (g *Gob) Delete(ctx context.Context, id string) (bool, error) {
	log := g.logger.Logger()

	// first refresh in-memory map
	err := g.MapRefresh(ctx)
	if err != nil {
		log.Error().Msgf("error while refresh gob persistent storage: error: %s\n", err.Error())
		return false, err
	}

	// delete id from map
	b, err := g.basin.Delete(ctx, id)
	if err != nil || !b {
		log.Debug().Msgf("error while deleting entry with id: %s : %s\n", id, err.Error())
		return false, err
	}

	// replace store with new map
	i, err := g.persist(ctx, true)
	if err != nil {
		log.Debug().Msgf("error while persisting entry with id: %s : %s\n", id, err.Error())
		return false, err
	}

	if i == 0 {
		return false, fmt.Errorf("wrote 0 bytes to gob persistent storage")
	}

	return true, nil
}

func (g *Gob) trunc(i int64) error {
	g.Lock()
	defer g.Unlock()
	err := g.fd.Truncate(i)
	if err != nil {
		log.Info().Msgf("error while emptying store: %s\n", err.Error())
		return err
	}
	return nil
}

func (g *Gob) persist(ctx context.Context, replace bool) (int64, error) {
	log := g.logger.Logger()

	// pre-checks and pre-reqs
	if replace {
		// first, truncate file to zero. clean the file.
		err := g.trunc(0)
		if err != nil {
			log.Debug().Msgf("error while cleaning gob persistent store: %s\n", err.Error())
			return 0, err
		}

	}

	// core persist
	// dump the current in-memory contents into file. TODO: return both the length of binary written for integrity checks, and any error if encountered.
	err := g.MapDump(ctx)
	if err != nil {
		log.Error().Msgf("error while dumping in-memory map : error: %s\n", err.Error())
		return 0, err
	}

	// post checks
	// compare the size of the file in bytes and the size of bytes written via MapDump from above
	f, err := g.fd.Stat()
	if err != nil {
		log.Error().Msgf("error retrieving file stat: error: %s\n", err.Error())
		return 0, err
	}

	return f.Size(), nil
}

// Close closes the redis connection
func (g *Gob) Close(ctx context.Context) error {
	return g.fd.Close()
}

// MapRefresh refreshes the sync.Map in-memory store with the latest updates from the persistent store
func (g *Gob) MapRefresh(ctx context.Context) error {
	// sets up a temporary store for the
	var m = map[string]string{}

	// lock
	g.RLock()
	defer g.RUnlock()

	//fBytes := []byte{}
	//g.fd.Read(fBytes)
	//
	//fBuffer := bytes.Buffer{}
	//_, err := fBuffer.Read(fBytes)
	//if err != nil {
	//	return err
	//}

	// set up gob decoder with file descriptor to gob persistent store
	dec := gob.NewDecoder(g.fd)

	// encode map and write to io.Writer
	err := dec.Decode(&m)
	if err == io.EOF {
		log.Warn().Msgf("decoding file returned with EOF: file empty: %s\n", err.Error())
	} else if err != nil {
		log.Error().Msgf("error while decoding into map from gob persistent storage: error: %s\n", err.Error())
		return fmt.Errorf("error while encoddecodinging into map from gob persistent storage: error: %s\n", err.Error())
	}

	// only perform a clean refresh when m map is not empty
	if len(m) > 0 {
		// first empty sync map
		// if any error is received from reading from file, the internal sync.Map isn't cleared
		_, err = g.basin.Flush(ctx)
		if err != nil {
			log.Error().Msgf("error flushing in-memory store: %s\n", err.Error())
			return fmt.Errorf("error flushing in-memory store: %s\n", err.Error())
		}

		// unfurl map into sync map
		syncM := g.basin.Map()
		for k, v := range m {
			syncM.Store(k, v)
		}
	}

	return nil
}

// MapDump persists the current in-memory data to the persistent store
func (g *Gob) MapDump(ctx context.Context) error {
	log := g.logger.Logger()

	g.Lock()
	defer g.Unlock()

	var m = map[string]string{}
	dec := gob.NewEncoder(g.fd)

	// store all the sync map contents into the temporary map
	g.basin.scaffold.Range(func(id, value interface{}) bool {
		m[fmt.Sprint(id)] = fmt.Sprint(value)
		return true
	})
	log.Debug().Msg("successfully ranged over sync.Map store")

	// print things about to be stored, for debugging purposes
	fmt.Println("about to persist:", m)

	// encode map and write to io.Writer || fd
	err := dec.Encode(m)
	if err != nil {
		log.Error().Msgf("error while encoding into map into gob persistent storage: error: %s\n", err.Error())
		return err
	}

	// reopen file to update internal file descriptor TODO: move this to a separate function later
	_ = g.fileRefresh()

	log.Info().Msg("successfully persisted in-memory map to disk")

	return nil
}

func (g *Gob) fileRefresh() error {
	var err error
	g.fd, err = os.OpenFile(g.fd.Name(), os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		fmt.Printf("error reading map: %s\n", err.Error())
		return err
	}
	return nil
}

// Flush empties the internal sync.Map and the persistent gob store // TODO: why? What is the use case for this?
// TODO: Flush should rather persist the current state of the in-memory map into disk, and then empty the in-memory map. It isn't idiomatic for flush to clear the persistent store too.
func (g *Gob) Flush(ctx context.Context) (bool, error) {
	log := g.logger.Logger()

	// first empty sync map
	b, err := g.basin.Flush(ctx)
	log.Debug().Msgf("flushing gob store")
	if !b || err != nil {
		log.Error().Msgf("error occurred while flushing gob store: %s\n", err.Error())
		return b, err
	}

	// delete persistent gob store last
	err = g.trunc(0)
	log.Debug().Msgf("flushing gob persistent store")
	if err != nil {
		log.Error().Msgf("error occurred while flushing persistent gob store: %s\n", err.Error())
		return false, err
	}
	return true, nil
}
