package infra

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync/atomic"
	"testrand-vm/config"
	"time"
)

type RemoteJointVariable struct {
	SessionId  string
	EtcdClient *clientv3.Client
}

func (env *RemoteJointVariable) GetClient() *clientv3.Client {
	return env.EtcdClient
}

func (env *RemoteJointVariable) Transaction(f func(stm concurrency.STM) error) (bool, error) {
	txn, err := concurrency.NewSTM(env.EtcdClient, func(stm concurrency.STM) error {
		return f(stm)
	})
	if err != nil {
		return false, err
	}
	return txn.Succeeded, err
}

func (env *RemoteJointVariable) Get(key string) (string, error) {
	r, err := env.EtcdClient.Get(context.Background(), key)
	if err != nil {
		return "", err
	}
	if len(r.Kvs) != 1 {
		return "", errors.New(fmt.Sprintf("not found key: %d", len(r.Kvs)))
	}
	return string(r.Kvs[0].Value), nil
}

func (env *RemoteJointVariable) GetAll() ([]struct {
	Key   string
	Value string
}, error) {
	r, err := env.EtcdClient.Get(context.TODO(), fmt.Sprintf("/env/%s", env.SessionId),
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	)
	if err != nil {
		return []struct {
			Key   string
			Value string
		}{}, err
	}
	var result []struct {
		Key   string
		Value string
	}
	for _, kv := range r.Kvs {
		result = append(result, struct {
			Key   string
			Value string
		}{Key: string(kv.Key), Value: string(kv.Value)})
	}
	return result, nil
}

func (env *RemoteJointVariable) Put(key string, value string, option clientv3.OpOption) error {
	if option == nil {
		_, err := env.EtcdClient.Put(context.Background(), key, value)
		return err
	}
	_, err := env.EtcdClient.Put(context.Background(), key, value, option)
	return err
}

func (env *RemoteJointVariable) ClearAll() error {
	_, err := env.EtcdClient.Delete(context.Background(), fmt.Sprintf("/env/%s", env.SessionId), clientv3.WithPrefix())
	return err
}

var locked atomic.Int32
var etcdClient *clientv3.Client

// setup etcd
func SetupEtcd(sessionId string) (*RemoteJointVariable, error) {
	conf := config.Get()

	var err error

	for locked.Load() != 0 {
	}
	locked.Add(1)
	if etcdClient != nil {
		locked.Add(-1)
		return &RemoteJointVariable{EtcdClient: etcdClient, SessionId: sessionId}, err
	}
	//setup etcd
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("http://%s:%s", conf.EtcdHost, conf.EtcdPort)},
		DialTimeout: 5 * time.Second,
	})
	locked.Add(-1)
	if err != nil {
		return nil, err
	}
	fmt.Println("etcd setup success")
	return &RemoteJointVariable{EtcdClient: etcdClient, SessionId: sessionId}, err
}
