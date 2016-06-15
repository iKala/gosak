package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"straas.io/pierce"
)

type Room interface {
	Start()
	Stop()
	Join(pierce.SocketConnection)
	Leave(pierce.SocketConnection)
	Empty() bool
}

func newRoom(id, etcdKey string, api client.KeysAPI) Room {
	log.Debugf("create room with key %s", etcdKey)

	return &roomImpl{
		id:        id,
		conns:     map[pierce.SocketConnection]bool{},
		connCount: 0,
		api:       api,
		key:       etcdKey,
		chJoin:    make(chan pierce.SocketConnection, 10),
		chLeave:   make(chan pierce.SocketConnection, 10),
		chDone:    make(chan bool),
	}
}

type roomImpl struct {
	id string
	// all pierce.SocketConnections in this room
	conns map[pierce.SocketConnection]bool
	// keep track real pierce.SocketConnection count (some might still in the channel)
	connCount int
	api       client.KeysAPI
	// channels
	chJoin  chan pierce.SocketConnection
	chLeave chan pierce.SocketConnection
	chDone  chan bool

	key     string
	data    map[string]interface{}
	dataStr string // cache data to avoid redundant marshalling
	version uint64
}

func (r *roomImpl) Start() {
	go r.mainLoop()
}

func (r *roomImpl) Stop() {
	close(r.chDone)
}

func (r *roomImpl) Empty() bool {
	return r.connCount == 0
}

func (r *roomImpl) Join(conn pierce.SocketConnection) {
	log.Infof("connection %s join %s", conn.Id(), r.id)
	r.connCount++
	r.chJoin <- conn
}

func (r *roomImpl) Leave(conn pierce.SocketConnection) {
	log.Infof("connection %s leave %s", conn.Id(), r.id)
	r.connCount--
	r.chLeave <- conn
}

func (r *roomImpl) join(conn pierce.SocketConnection) {
	r.conns[conn] = true
	log.Infof("there %d conns in room %s", len(r.conns), r.id)

	// send if has data
	if r.version > 0 {
		conn.Emit(r.dataStr, r.version)
	}
}

func (r *roomImpl) leave(conn pierce.SocketConnection) {
	// TODO: furthor notification ?!
	delete(r.conns, conn)
}

func (r *roomImpl) mainLoop() {
	// how to make sure alreay watching ?!
	wch := r.watch()
	for {
		select {
		case conn := <-r.chJoin:
			r.join(conn)

		case conn := <-r.chLeave:
			r.leave(conn)

		case resp := <-wch:
			// TODO: in another goroutine ?!
			if err := r.applyChange(resp); err != nil {
				log.Errorf("fail to apply resp %+v, err:%v", resp, err)
				// WTD
				continue
			}
			r.broadcast()
		}
	}
}

func (r *roomImpl) watch() <-chan *client.Response {

	// get value recursively
	getValue := func() (*client.Response, error) {
		opt := &client.GetOptions{
			Recursive: true,
			// TODO: not sure quorum is necessary or not
			// Quorum:    true,
		}

		// TODO: context with timeout
		ctx := context.Background()
		resp, err := r.api.Get(ctx, r.key, opt)
		if cErr, ok := err.(client.Error); ok {
			if cErr.Code != client.ErrorCodeKeyNotFound {
				return resp, err
			}
			// simulate an empty dir
			return &client.Response{
				Action: "get",
				Index:  cErr.Index,
				Node: &client.Node{
					Dir: true,
					Key: r.key,
				},
			}, nil
		}
		return resp, err

	}
	// watch key recursively
	getWatcher := func(index uint64) client.Watcher {
		opt := &client.WatcherOptions{
			Recursive:  true,
			AfterIndex: index,
		}
		return r.api.Watcher(r.key, opt)
	}

	ch := make(chan *client.Response, 10)
	go func() {
		log.Infof("start to watch for room")
		for {
			resp, err := getValue()
			if err != nil {
				log.Errorf("fail to get value, err:%v", err)
				// TODO: key not found
				// TODO: backoff
				continue
			}
			// check return ?
			select {
			case <-r.chDone:
				// terminate watcher loop
				return
			default:
			}
			ch <- resp

			// TODO: index plus one ?!
			w := getWatcher(resp.Index)
			for {
				ctx, cancel := context.WithCancel(context.Background())
				// must use goroutine in order to cancel watch
				go func() {
					resp, err = w.Next(ctx)
					// it's ok to call cancel multiple times
					cancel()
				}()

				select {
				case <-r.chDone:
					cancel()
					// terminate watcher loop
					return
				case <-ctx.Done():
				}

				if err != nil {
					cerr, ok := err.(*client.Error)
					if ok && cerr.Code == client.ErrorCodeEventIndexCleared {
						break
					}
					fmt.Println(err)
					// backoff if other error ?!
					continue
				}
				ch <- resp
			}
		}
	}()

	return ch
}

func (r *roomImpl) applyChange(resp *client.Response) error {
	fmt.Printf("%+v\n", resp)
	cur := resp.Node

	// get room and key from etcd key
	key, err := getKey(r.key, cur.Key)
	if err != nil {
		// illegal key
		return err
	}

	data, version, err := toValue(cur)
	if err != nil {
		// WTF
		return err
	}
	// older changes, just ignore it
	if version <= r.version {
		// TODO: log
		return nil
	}
	r.version = version

	switch resp.Action {
	case "get":
		r.data = data

	case "create", "set", "update":
		if key == "" {
			r.data = data
		} else {
			r.data[key] = data
		}

	case "delete", "expire":
		if key == "" {
			r.data = map[string]interface{}{}
		} else {
			delete(r.data, key)
		}

	default:
		// should not reach here
		// TODO: keep log and metrics
	}
	return nil
}

func (r *roomImpl) broadcast() {
	bs, _ := json.Marshal(r.data)
	r.dataStr = string(bs)

	// TODO: check previous value
	for conn := range r.conns {
		conn.Emit(r.dataStr, r.version)
	}
}

// toValue performs DFS to build value map and max modified idx
func toValue(node *client.Node) (map[string]interface{}, uint64, error) {
	if node.Dir {
		vs := map[string]interface{}{}
		maxIndex := node.ModifiedIndex
		for _, n := range node.Nodes {
			key, err := getKey(node.Key, n.Key)
			if err != nil {
				return nil, 0, err
			}
			log.Infof("forge key %s %s => %s", node.Key, n.Key, key)

			v, idx, err := toValue(n)
			if err != nil {
				return nil, 0, err
			}
			if idx > maxIndex {
				maxIndex = idx
			}
			vs[key] = v
		}
		return vs, maxIndex, nil
	} else {
		if node.Value == "" {
			return nil, node.ModifiedIndex, nil
		}
		log.Infof("%s=%s", node.Key, node.Value)
		v := map[string]interface{}{}
		if err := json.Unmarshal([]byte(node.Value), &v); err != nil {
			return nil, 0, err
		}
		return v, node.ModifiedIndex, nil
	}
}

func getKey(prefix, etcdKey string) (string, error) {
	if !strings.HasPrefix(etcdKey, prefix) {
		return "", fmt.Errorf("unexcepted etcd key %s for %s", etcdKey, prefix)
	}
	return strings.TrimPrefix(etcdKey[len(prefix):], "/"), nil
}
