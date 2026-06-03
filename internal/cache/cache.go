package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
	"github.com/adrg/xdg"
)

type Cache[T any] struct {
	format string
	prefix string
	ttl    time.Duration
}

type Config[T any] struct {
	Format string
	Prefix string
	TTL    time.Duration
}

func New[T any](conf *Config[T]) *Cache[T] {
	return &Cache[T]{
		format: conf.Format,
		prefix: conf.Prefix,
		ttl:    conf.TTL,
	}
}

func (c *Cache[T]) path(key string) (string, error) {
	filename := c.prefix
	if key != "" {
		filename = fmt.Sprintf("%s_%s", c.prefix, key)
	}
	return xdg.CacheFile(filepath.Join(config.ProjectName, filename+".json"))
}

func (c *Cache[T]) encode(data *T) ([]byte, error) {
	switch c.format {
	case "json":
		return json.Marshal(data)
	default:
		return nil, fmt.Errorf("unsupported cache format: %s", c.format)
	}
}

func (c *Cache[T]) decode(blob []byte) (*T, error) {
	var data T
	var err error
	switch c.format {
	case "json":
		err = json.Unmarshal(blob, &data)
	default:
		return nil, fmt.Errorf("unsupported cache format: %s", c.format)
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func Key(parts ...string) string {
	return strings.Join(parts, "_")
}

func (c *Cache[T]) isValid(key string) (bool, error) {
	path, err := c.path(key)
	if err != nil {
		return false, err
	}
	return util.IsFileModifiedWithin(path, c.ttl), nil
}

func (c *Cache[T]) Get(key string) (*T, error) {
	path, err := c.path(key)
	if err != nil {
		return nil, err
	}

	if valid, err := c.isValid(key); err != nil || !valid {
		return nil, errors.Join(err, c.Delete(key))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	return c.decode(data)
}

func (c *Cache[T]) Set(key string, data T) error {
	path, err := c.path(key)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	blob, err := c.encode(&data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, blob, 0644)
}

func (c *Cache[T]) Delete(key string) error {
	path, err := c.path(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
