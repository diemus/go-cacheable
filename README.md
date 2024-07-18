# go-cacheable

[![Go Report Card](https://goreportcard.com/badge/github.com/diemus/go-cacheable)](https://goreportcard.com/report/github.com/diemus/go-cacheable)
[![License](https://badgen.net/badge/license/MIT/cyan)](https://github.com/diemus/go-cacheable/blob/main/LICENSE)
[![Release](https://badgen.net/github/release/diemus/go-cacheable/latest)](https://github.com/diemus/go-cacheable)
[![Azure](https://badgen.net/badge/icon/Golang?icon=azure&label)](https://github.com/diemus/go-cacheable)

## Introduction

<a href="./README.md">English</a> |
<a href="./README.zh-cn.md">中文</a>

go-cacheable is a function wrapper that wraps arbitrary functions with advanced cache management. Inspired by Java Spring Cacheable and Golang singleflight, this library provides a concise API for managing and operating caches. It allows quick addition of caching functionality to functions, supports multiple storage backends, and comes with advanced features such as cache penetration prevention, tag management, and Prometheus monitoring metrics.

## Highlights

1. ✅ **Easy Integration**: Simple API design, consistent with Golang design philosophy, easy to integrate and use in existing projects.
2. ✅ **Type Safety**: Utilizes Go's generic features for automatic inference, no type conversion needed.
3. ✅ **Cache Penetration Prevention**: Built-in singleflight mechanism effectively prevents cache penetration issues.
4. ✅ **Flexible Caching Strategies**: Supports custom expiration times and tag-based expiration strategies to meet various caching needs.
5. ✅ **Observability**: Built-in Prometheus metrics collection for easy monitoring and debugging.
6. ✅ **Multiple Cache Backends**: Supports various popular cache storage backends, flexibly adapting to different application scenarios.

## Installation

Install go-cacheable using go get:

```bash
go get github.com/diemus/go-cacheable
```

## Quick Start

### Initialize Cache Manager

First, you need to initialize the cache manager. go-cacheable supports multiple storage backends. Here we'll use Redis and local memory cache as examples. In practice, you can choose which cache managers to initialize based on your needs:

```go
import (
    "github.com/diemus/go-cacheable"
    "github.com/eko/gocache/lib/v4/store"
    "github.com/eko/gocache/lib/v4/store/rediscluster"
    "github.com/eko/gocache/lib/v4/store/go_cache"
    "github.com/redis/go-redis/v9"
    "github.com/patrickmn/go-cache"
    "time"
)

var RemoteCacheManager *cacheable.CacheManager
var LocalCacheManager *cacheable.CacheManager

func InitCacheManager() {
    // Initialize Redis client
    redisClient := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{":6379"},
    })
    redisStore := rediscluster.NewRedisCluster(redisClient)

    // Initialize local cache
    goCacheClient := cache.New(5*time.Minute, 10*time.Minute)
    goCacheStore := go_cache.NewGoCache(goCacheClient)

    // Create cache managers
    RemoteCacheManager = cacheable.NewCacheManager(redisStore)
    LocalCacheManager = cacheable.NewCacheManager(goCacheStore)

    // Set global configurations (optional)
    cacheable.SetDefaultKeyPrefix("myapp")
    cacheable.SetDefaultExpiration(5 * time.Minute)
}
```

### Basic Usage

Use the `Get` function to wrap a function and add caching capability:

Before modification:
```go
func GetUser(ctx context.Context, id int) (User, error) {
    return fetchUserFromDatabase(id)
}
```

After modification, due to the use of generics, the return value is consistent with the original function:
```go
func GetUser(ctx context.Context, id int) (User, error) {
    user, err, _ := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
        // This is the function to get user information when cache miss
        return fetchUserFromDatabase(id)
    })
    return user, err
}
```

Without modifying the original function, cache only when calling, suitable for situations where sometimes caching is needed and sometimes direct database query is needed:
```go
user, err, cached := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
    // This is the function to get user information when cache miss
    return GetUser(id)
})
```

The Get function can flexibly specify whether to use remote cache or local cache:
```go
func GetUser(ctx context.Context, id int) (User, error) {
    user, err, _ := cacheable.Get(ctx, LocalCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
        // This is the function to get user information when cache miss
        return fetchUserFromDatabase(id)
    })
    return user, err
}
```

### Using Options

go-cacheable provides multiple options to customize caching behavior:

```go
user, err, _ := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), fetchUserFromDatabase,
    cacheable.WithExpiration(10 * time.Minute),
    cacheable.WithTags("user", fmt.Sprintf("teamId:%d", user.TeamId)),
    cacheable.WithDynamicTags(func() []string {
        // This function is only called when setting the cache
        teamIDs, _ := getTeamIDsForUser("user1")
        tags := make([]string, len(teamIDs))
        for i, id := range teamIDs {
            tags[i] = fmt.Sprintf("teamId:%d", id)
        }
        return tags
    }),
)
```

### Purpose of Tags

Tags are used to define metadata for caches, facilitating batch deletion. For example, if the cache key is username, the tag can be teamId. When a team changes, all user caches related to that team can be deleted:

```go
// Add tag when setting cache
value, err, _ := cacheable.Get(ctx, cacheManager, "users", username, getUserData,
    cacheable.WithTags(fmt.Sprintf("teamId:%d", userTeamID)),
)

// Delete related caches when team changes
err := cacheable.DeleteByTags(ctx, cacheManager, []string{fmt.Sprintf("teamId:%d", changedTeamID)})
```

### Dynamic Tags

The purpose of dynamic tags is to handle scenarios where computing the tag might also be a time-consuming operation. For example, finding a user's teamId might require a database query. With dynamic tags, this computation only occurs when setting the cache, not every time the cache is accessed:

```go
value, err, cached := cacheable.Get(ctx, cacheManager, "users", "user1", getUserData,
    cacheable.WithDynamicTags(func() []string {
        // This function is only called when setting the cache
        teamIDs, _ := getTeamIDsForUser("user1")
        tags := make([]string, len(teamIDs))
        for i, id := range teamIDs {
            tags[i] = fmt.Sprintf("teamId:%d", id)
        }
        return tags
    }),
)
```

### Deleting Cache

Delete a single cache item:

```go
err := cacheable.Delete(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id))
```

Delete cache based on tags, for example, clear all team member caches when a team changes:

```go
err := cacheable.DeleteByTags(ctx, RemoteCacheManager, []string{"teamId:123"})
```
### Multiple Cache Backends

go-cacheable is built on top of [github.com/eko/gocache](https://github.com/eko/gocache) and supports multiple cache backends:

```go
// Install required backend stores
go get github.com/eko/gocache/store/bigcache/v4
go get github.com/eko/gocache/store/freecache/v4
go get github.com/eko/gocache/store/go_cache/v4
go get github.com/eko/gocache/store/hazelcast/v4
go get github.com/eko/gocache/store/memcache/v4
go get github.com/eko/gocache/store/pegasus/v4
go get github.com/eko/gocache/store/redis/v4
go get github.com/eko/gocache/store/rediscluster/v4
go get github.com/eko/gocache/store/rueidis/v4
go get github.com/eko/gocache/store/ristretto/v4

// Usage example (using Redis)
import "github.com/eko/gocache/store/redis/v4"

redisStore := redis.NewRedis(redisClient)
cacheManager := cacheable.NewCacheManager(redisStore)
```

## Metrics Collection

Go-Cacheable has built-in metrics collection functionality that can be easily integrated into your monitoring system:

```go
// Metrics are collected automatically, you can view the following metrics in your metrics system:
cacheable_cache_request_total{namespace="xxx"}
cacheable_cache_hit_total{namespace="xxx"}
```

## Configuration

You can set global default values using the following methods:

```go
cacheable.SetDefaultKeyPrefix("cacheable")
cacheable.SetDefaultExpiration(30 * time.Minute)
cacheable.SetDefaultMetricsPrefix("cacheable")
```

## License

MIT

## Project Popularity

[![Star History Chart](https://api.star-history.com/svg?repos=diemus/go-cacheable&type=Date)](https://star-history.com/#diemus/go-cacheable&Date)