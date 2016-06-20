package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"straas.io/base/logger"
	"straas.io/external"
)

const (
	bufferSize = 10
)

// NewEtcd creates an instance of etcd
func NewEtcd(c client.Client, timeout time.Duration, log logger.Logger) external.Etcd {
	return &etcdImpl{
		api:     &keysImpl{api: client.NewKeysAPI(c)},
		timeout: timeout,
		log:     log,
	}
}

// keysAPI bridge etc.KeysAPI interface for testing purpose
// bcz we use vendor and etc/client use vendor as well, keysAPI mock implmenetation in our pkg
// will lead to type mismatch problem:
//   have Get("vendor/github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context".Context ...
//   want Get("vendor/golang.org/x/net/context".Context ...
// golang.org/x/net/context will be treated as different type
// context will be build-in pkg in golang 1.7, then we cloud get rid of this problem
type keysAPI interface {
	// Get bridges etcd Get
	Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error)
	// Set bridges etcd Set
	Set(ctx context.Context, key, value string, opts *client.SetOptions) (*client.Response, error)
	// Watcher bridges etcd Watcher
	Watcher(key string, opts *client.WatcherOptions) watcher
}

// watcher is also for bridging etc.Watcher
type watcher interface {
	Next(ctx context.Context) (*client.Response, error)
}

type etcdImpl struct {
	log     logger.Logger
	api     keysAPI
	timeout time.Duration
}

// GetAndWatch returns a chan for etcd response, this function will handle error reconnect
// and outdate.
func (a *etcdImpl) GetAndWatch(etcdKey string, done <-chan bool) <-chan *client.Response {
	// check if need to leave loop
	checkDone := func() bool {
		// check return ?
		select {
		case <-done:
			// need to terminate watcher loop
			return true
		default:
			return false
		}
	}

	ch := make(chan *client.Response, bufferSize)
	log := a.log

	go func() {
		for {
			if checkDone() {
				return
			}

			resp, err := a.Get(etcdKey, true)
			if err != nil {
				log.Errorf("fail to get value, err:%v", err)
				// TODO: backoff
				continue
			}
			ch <- resp

			// TODO: index plus one ?!
			w := a.getWatcher(etcdKey, resp.Index)
			for {
				ctx, cancel := context.WithCancel(context.Background())
				// must use goroutine in order to cancel watch
				go func() {
					resp, err = w.Next(ctx)
					// it's ok to call cancel multiple times
					cancel()
				}()

				select {
				case <-done:
					cancel()
					// terminate watcher loop
					return
				case <-ctx.Done():
				}

				if err != nil {
					cerr, ok := err.(*client.Error)
					// index outdate, need to restart watch loop
					if ok && cerr.Code == client.ErrorCodeEventIndexCleared {
						break
					}

					// What to do ?
					log.Errorf("fail to watch, err:%v", err)
					// TODO: backoff if other error ?!
					continue
				}
				ch <- resp
			}
		}
	}()

	return ch
}

func (a *etcdImpl) Get(etcdKey string, recursive bool) (*client.Response, error) {
	opt := &client.GetOptions{
		Recursive: recursive,
		// TODO: not sure quorum is necessary or not
		// Quorum:    true,
	}
	ctx, _ := context.WithTimeout(context.Background(), a.timeout)
	resp, err := a.api.Get(ctx, etcdKey, opt)
	if cErr, ok := err.(client.Error); ok && cErr.Code == client.ErrorCodeKeyNotFound {
		// simulate an empty dir
		return emptyDirResponse(etcdKey, cErr.Index), nil
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *etcdImpl) Set(etcdKey, value string) (*client.Response, error) {
	opt := &client.SetOptions{}
	ctx, _ := context.WithTimeout(context.Background(), a.timeout)
	return a.api.Set(ctx, etcdKey, value, opt)
}

// getWatcher returns watcher with the given key and index
func (a *etcdImpl) getWatcher(etcdKey string, index uint64) watcher {
	opt := &client.WatcherOptions{
		Recursive:  true,
		AfterIndex: index,
	}
	return a.api.Watcher(etcdKey, opt)
}

// emptyDirResponse return a etcd dir get response
func emptyDirResponse(etcdKey string, index uint64) *client.Response {
	return &client.Response{
		Action: "get",
		Index:  index,
		Node: &client.Node{
			Dir: true,
			Key: etcdKey,
		},
	}
}
