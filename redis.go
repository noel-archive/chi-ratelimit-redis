// 🔬 chi-ratelimit-redis: Redis support for the chi-ratelimit library.
// Copyright (c) 2022 Noelware
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package redis is the main package to use a Redis server to persist
// ratelimits.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/noelware/chi-ratelimit/providers"
	"github.com/noelware/chi-ratelimit/types"
	"time"
)

// Provider is the main providers.Provider object to implement when using
// this library.
type Provider struct {
	keyPrefix string
	client    *redis.Client
}

type options struct {
	keyPrefix string
	client    *redis.Client
}

// WithKeyPrefix appends a new key prefix to use when constructing
// a Provider.
func WithKeyPrefix(prefix string) func(o *options) {
	return func(o *options) {
		o.keyPrefix = prefix
	}
}

// WithClient appends a pre-existing Redis client that is connected
// when constructing a Provider.
func WithClient(client *redis.Client) func(o *options) {
	return func(o *options) {
		o.client = client
	}
}

// WithConfig creates and connects a new Redis client and appends it
// to the Provider.
func WithConfig(config *redis.Options) (func(o *options), error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	client := redis.NewClient(config)
	if err := client.Ping(ctx).Err(); err != nil {
		// TODO: find a better solution for this
		// no-op operation
		return func(o *options) {}, nil
	}

	return func(o *options) {
		o.client = client
	}, nil
}

// New creates a new Provider object with the following options that was
// passed down.
func New(opts ...func(o *options)) (providers.Provider, error) {
	config := &options{
		keyPrefix: "chi_ratelimit",
		client:    nil,
	}

	for _, override := range opts {
		override(config)
	}

	if config.client == nil {
		return nil, errors.New("missing redis client to use")
	}

	return &Provider{
		keyPrefix: config.keyPrefix,
		client:    config.client,
	}, nil
}

func (p *Provider) Reset(key string) (bool, error) {
	// Check if it exists
	ok, err := p.client.HExists(context.TODO(), p.keyPrefix, key).Result()
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	// Delete it from Redis
	if err := p.client.HDel(context.TODO(), p.keyPrefix, key).Err(); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (*Provider) Name() string {
	return "redis provider"
}

func (p *Provider) Put(key string, value *types.Ratelimit) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := p.client.HMSet(context.TODO(), p.keyPrefix, key, string(data)).Err(); err != nil {
		return err
	} else {
		return nil
	}
}

func (p *Provider) Get(key string) (*types.Ratelimit, error) {
	// Update the database with the new copy
	data, err := p.client.HGet(context.TODO(), p.keyPrefix, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var rl *types.Ratelimit
	if err := json.Unmarshal([]byte(data), &rl); err != nil {
		return nil, err
	}

	copied := rl.Copy()
	if err := p.Put(key, copied); err != nil {
		return nil, err
	}

	return copied, nil
}
