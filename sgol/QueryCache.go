package sgol

import (
  "time"
  "github.com/patrickmn/go-cache"
)

type QueryCache struct {
  Enabled bool `json:"enabled" hcl:"enabled"`
  DefaultExpiration int `json:"default_expiration" hcl:"default_expiration"`
  CleanupInterval int `json:"cleanup_interval" hcl:"cleanup_interval"`
  ExcludeEntities []string `json:"exclude_entities" hcl:"exclude_entities"`
  ExcludeEdges []string `json:"exclude_edges" hcl:"exclude_edges"`
  Cache *cache.Cache
}

type QueryCacheInstance struct {
  DefaultExpiration int `json:"default_expiration" hcl:"default_expiration"`
  CleanupInterval int `json:"cleanup_interval" hcl:"cleanup_interval"`
  ExcludeEntities []string `json:"exclude_entities" hcl:"exclude_entities"`
  ExcludeEdges []string `json:"exclude_edges" hcl:"exclude_edges"`
  Cache *cache.Cache
}

func NewQueryCache(c *QueryCache) (*QueryCacheInstance) {
  qci := &QueryCacheInstance{
    DefaultExpiration: c.DefaultExpiration,
    CleanupInterval: c.CleanupInterval,
    ExcludeEntities: c.ExcludeEntities,
    ExcludeEdges: c.ExcludeEdges,
  }
  qci.Cache = cache.New(
    time.Duration(qci.DefaultExpiration)*time.Minute,
    time.Duration(qci.CleanupInterval)*time.Minute,
  )
  return qci
}

func (c *QueryCacheInstance) CacheOperationChain(chain OperationChain) (bool) {
  if chain.Limit > 0 {
    return false
  }

  for _, op := range chain.Operations {
    switch op.(type) {
    case OperationSelect:
      for _, entity := range op.(OperationSelect).Entities {
        if StringSliceContains(c.ExcludeEntities, entity) {
          return false
        }
      }
    case OperationNav:
      for _, entity := range op.(OperationNav).Entities {
        if StringSliceContains(c.ExcludeEntities, entity) {
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

func (c *QueryCacheInstance) GetOperationChain(chain OperationChain) (interface{}, bool, error) {
  key, err := chain.Hash()
  if err != nil {
    return "", false, err
  }
  value, found := c.Cache.Get(key)
  return value, found, nil
}

func (c *QueryCacheInstance) Set(key string, value interface{}) {
  c.Cache.Set(key, value, cache.NoExpiration)
}

func (c *QueryCacheInstance) SetOperationChain(chain OperationChain, value interface{}) (error) {
  key, err := chain.Hash()
  if err != nil {
    return err
  }
  c.Cache.Set(key, value, cache.NoExpiration)
  return nil
}

func (c *QueryCacheInstance) SetWithExpiration(key string, value interface{}, duration time.Duration) {
  c.Cache.Set(key, value, duration)
}

func (c *QueryCacheInstance) Validate(chain *OperationChain) bool {
  if chain.Limit > 0 {
    return false
  }
  return true
}
