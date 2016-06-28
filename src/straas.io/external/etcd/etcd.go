package etcd

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"straas.io/base/logmetric"
	"straas.io/external"
)

const (
	bufferSize = 10
)

// NewEtcd creates an instance of etcd
func NewEtcd(c client.Client, timeout time.Duration, logm logmetric.LogMetric) external.Etcd {
	return &etcdImpl{
		api:     &keysImpl{api: client.NewKeysAPI(c)},
		timeout: timeout,
		logm:    logm,
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
	logm    logmetric.LogMetric
	api     keysAPI
	timeout time.Duration
}

func (a *etcdImpl) Watch(etcdKey string, afterIndex uint64,
	chResp chan<- *client.Response, done <-chan bool) *client.Error {
	logm := a.logm
	w := a.getWatcher(etcdKey, afterIndex)

	for {
		var resp *client.Response
		var err error

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
			return nil
		case <-ctx.Done():
		}

		if err != nil {
			cerr, ok := err.(*client.Error)
			// index outdate, need to restart watch loop
			// TODO: check client.EcodeWatcherCleared
			if ok && cerr.Code == client.ErrorCodeEventIndexCleared {
				logm.BumpSum("etcd.indexoutdated.err", 1)
				return cerr
			}

			// What to do ?
			logm.BumpSum("etcd.watchnext.err", 1)
			logm.Errorf("fail to watch, err:%v", err)
			// TODO: backoff if other error ?!
			continue
		}
		chResp <- resp
	}
}

// GetAndWatch returns a chan for etcd response, this function will handle error reconnect
// and outdate.
func (a *etcdImpl) GetAndWatch(etcdKey string, chResp chan<- *client.Response, done <-chan bool) {
	logm := a.logm
	for {
		// check if need to leave loop
		select {
		case <-done:
			// need to terminate watcher loop
			return
		default:
		}

		resp, err := a.Get(etcdKey, true)
		if err != nil {
			logm.Errorf("fail to get value, err:%v", err)
			// TODO: backoff
			continue
		}
		chResp <- resp

		// TODO: index plus one ?!
		if err := a.Watch(etcdKey, resp.Index, chResp, done); err != nil {
			// critical erro
		}
	}
}

func (a *etcdImpl) Get(etcdKey string, recursive bool) (*client.Response, error) {
	logm := a.logm
	logm.BumpSum("etcd.get", 1)
	defer logm.BumpTime("etcd.get.proc_time").End()

	opt := &client.GetOptions{
		Recursive: recursive,
		// TODO: not sure quorum is necessary or not
		// Quorum:    true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	resp, err := a.api.Get(ctx, etcdKey, opt)
	if cErr, ok := err.(client.Error); ok && cErr.Code == client.ErrorCodeKeyNotFound {
		// simulate an empty dir
		return emptyDirResponse(etcdKey, cErr.Index), nil
	}
	if err != nil {
		logm.BumpSum("etcd.get.err", 1)
		return nil, err
	}
	return resp, nil
}

func (a *etcdImpl) Set(etcdKey, value string) (*client.Response, error) {
	return a.SetWithTTL(etcdKey, value, 0)
}

func (a *etcdImpl) SetWithTTL(etcdKey, value string, ttl time.Duration) (*client.Response, error) {
	logm := a.logm
	logm.BumpSum("etcd.set", 1)
	defer logm.BumpTime("etcd.set.proc_time").End()

	opt := &client.SetOptions{}
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	// give ttl if not zero
	if ttl > 0 {
		opt.TTL = ttl
	}

	resp, err := a.api.Set(ctx, etcdKey, value, opt)
	if err != nil {
		logm.BumpSum("etcd.set.err", 1)
		return nil, err
	}
	return resp, nil
}

func (a *etcdImpl) RefreshTTL(etcdKey string, ttl time.Duration) (*client.Response, error) {
	logm := a.logm
	logm.BumpSum("etcd.refresh", 1)
	defer logm.BumpTime("etcd.refresh.proc_time").End()

	if ttl == 0 {
		return nil, fmt.Errorf("refresh etcd key %s with zero ttl", etcdKey)
	}

	opt := &client.SetOptions{
		Refresh:   true,
		PrevExist: client.PrevExist,
		TTL:       ttl,
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	resp, err := a.api.Set(ctx, etcdKey, "", opt)
	if err != nil {
		logm.BumpSum("etcd.refresh.err", 1)
		return nil, err
	}
	return resp, nil
}

func (a *etcdImpl) IsNotFound(err error) bool {
	return client.IsKeyNotFound(err)
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
