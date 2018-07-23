package sgol

import (
	"github.com/patrickmn/go-cache"
	"time"
)

import (
	"github.com/spatialcurrent/go-graph/graph"
	"github.com/spatialcurrent/sgol-codec/codec"
)

type QueryCache struct {
	Enabled           bool     `json:"enabled" hcl:"enabled"`
	DefaultExpiration int      `json:"default_expiration" hcl:"default_expiration"`
	CleanupInterval   int      `json:"cleanup_interval" hcl:"cleanup_interval"`
	ExcludeEntities   []string `json:"exclude_entities" hcl:"exclude_entities"`
	ExcludeEdges      []string `json:"exclude_edges" hcl:"exclude_edges"`
}

type QueryCacheInstance struct {
	DefaultExpiration int      `json:"default_expiration" hcl:"default_expiration"`
	CleanupInterval   int      `json:"cleanup_interval" hcl:"cleanup_interval"`
	ExcludeEntities   []string `json:"exclude_entities" hcl:"exclude_entities"`
	ExcludeEdges      []string `json:"exclude_edges" hcl:"exclude_edges"`
	Cache             *cache.Cache
}

func NewQueryCache(c *QueryCache) *QueryCacheInstance {
	qci := &QueryCacheInstance{
		DefaultExpiration: c.DefaultExpiration,
		CleanupInterval:   c.CleanupInterval,
		ExcludeEntities:   c.ExcludeEntities,
		ExcludeEdges:      c.ExcludeEdges,
	}
	qci.Cache = cache.New(
		time.Duration(qci.DefaultExpiration)*time.Minute,
		time.Duration(qci.CleanupInterval)*time.Minute,
	)
	return qci
}

func (c *QueryCacheInstance) CacheOperationChain(chain graph.OperationChain) bool {

	if chain.GetLimit() > 0 {
		return false
	}

	for _, op := range chain.GetOperations() {
		switch op.GetTypeName() {
		case "SELECT":
			op_select := op.(codec.OperationSelect)
			for _, g := range c.ExcludeEntities {
				if op_select.HasGroup(g) {
					return false
				}
			}
		case "NAV":
			op_nav := op.(codec.OperationNav)
			for _, g := range c.ExcludeEntities {
				if op_nav.HasEntityGroup(g) {
					return false
				}
			}
			for _, g := range c.ExcludeEdges {
				if op_nav.HasEdgeGroup(g) {
					return false
				}
			}
		case "HAS":
			op_has := op.(codec.OperationHas)
			for _, g := range c.ExcludeEntities {
				if op_has.HasEntityGroup(g) {
					return false
				}
			}
			for _, g := range c.ExcludeEdges {
				if op_has.HasEdgeGroup(g) {
					return false
				}
			}
		}
	}

	return true
}

func (c *QueryCacheInstance) Get(key string) (interface{}, bool) {
	return c.Cache.Get(key)
}

func (c *QueryCacheInstance) GetOperationChain(chain graph.OperationChain) (graph.QueryResponse, bool, error) {
	qr := graph.NewQueryResponse(false, "", "")

	key, err := chain.Hash()
	if err != nil {
		return qr, false, err
	}

	value, found := c.Cache.Get(key)
	if found {
		qr, err = graph.ParseQueryResponse(value.(string))
		if err != nil {
			return qr, found, err
		}
	}

	return qr, found, nil
}

func (c *QueryCacheInstance) GetHash(hash string) (graph.QueryResponse, bool, error) {
	qr := graph.NewQueryResponse(false, "", "")

	value, found := c.Cache.Get(hash)
	if !found {
		return graph.NewQueryResponse(false, "", ""), false, nil
	}

	qr, err := graph.ParseQueryResponse(value.(string))
	if err != nil {
		return qr, true, err
	}

	return qr, true, nil
}

func (c *QueryCacheInstance) Set(key string, value interface{}) {
	c.Cache.Set(key, value, cache.NoExpiration)
}

func (c *QueryCacheInstance) SetOperationChain(chain graph.OperationChain, qr graph.QueryResponse) error {
	hash, err := chain.Hash()
	if err != nil {
		return err
	}

	return c.SetHash(hash, qr)
}

func (c *QueryCacheInstance) SetHash(hash string, qr graph.QueryResponse) error {
	value, err := qr.Json()
	if err != nil {
		return err
	}
	c.Cache.Set(hash, value, cache.NoExpiration)
	return nil
}

func (c *QueryCacheInstance) SetWithExpiration(key string, value interface{}, duration time.Duration) {
	c.Cache.Set(key, value, duration)
}

func (c *QueryCacheInstance) Validate(chain graph.OperationChain) bool {
	if chain.GetLimit() > 0 {
		return false
	}
	return true
}
