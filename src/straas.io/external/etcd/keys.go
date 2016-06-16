package etcd

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type keysImpl struct {
	api client.KeysAPI
}

func (k *keysImpl) Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error) {
	return k.api.Get(ctx, key, opts)
}

func (k *keysImpl) Set(ctx context.Context, key, value string, opts *client.SetOptions) (*client.Response, error) {
	return k.api.Set(ctx, key, value, opts)
}

func (k *keysImpl) Watcher(key string, opts *client.WatcherOptions) watcher {
	w := k.api.Watcher(key, opts)
	return &watcherImpl{watcher: w}
}

type watcherImpl struct {
	watcher client.Watcher
}

func (w *watcherImpl) Next(ctx context.Context) (*client.Response, error) {
	return w.watcher.Next(ctx)
}
