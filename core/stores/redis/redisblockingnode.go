package redis

import (
	"fmt"
	"github.com/meta-apex/zenith/zlog"

	red "github.com/redis/go-redis/v9"
)

// ClosableNode interface represents a closable redis node.
type ClosableNode interface {
	RedisNode
	Close()
}

// CreateBlockingNode returns a ClosableNode.
func CreateBlockingNode(r *Redis) (ClosableNode, error) {
	timeout := readWriteTimeout + blockingQueryTimeout

	switch r.Type {
	case NodeType:
		client := red.NewClient(&red.Options{
			Addr:         r.Addr,
			Username:     r.User,
			Password:     r.Pass,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &clientBridge{client}, nil
	case ClusterType:
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        splitClusterAddrs(r.Addr),
			Username:     r.User,
			Password:     r.Pass,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &clusterBridge{client}, nil
	default:
		return nil, fmt.Errorf("unknown redis type: %s", r.Type)
	}
}

type (
	clientBridge struct {
		*red.Client
	}

	clusterBridge struct {
		*red.ClusterClient
	}
)

func (bridge *clientBridge) Close() {
	if err := bridge.Client.Close(); err != nil {
		zlog.Error().Msgf("Error occurred on close redis client: %s", err)
	}
}

func (bridge *clusterBridge) Close() {
	if err := bridge.ClusterClient.Close(); err != nil {
		zlog.Error().Msgf("Error occurred on close redis cluster: %s", err)
	}
}
