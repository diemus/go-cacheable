# go-cacheable

[![Go Report Card](https://goreportcard.com/badge/github.com/diemus/go-cacheable)](https://goreportcard.com/report/github.com/diemus/go-cacheable)
[![License](https://badgen.net/badge/license/MIT/cyan)](https://github.com/diemus/go-cacheable/blob/main/LICENSE)
[![Release](https://badgen.net/github/release/diemus/go-cacheable/latest)](https://github.com/diemus/go-cacheable)
[![Azure](https://badgen.net/badge/icon/Golang?icon=azure&label)](https://github.com/diemus/go-cacheable)

## 介绍

<a href="./README.md">English</a> |
<a href="./README.zh-cn.md">中文</a>

go-cacheable 是一个函数包装器，用于将任意函数包装为带有高级缓存管理的函数，此库受Java Spring Cacheable与Golang singlefilght启发，提供了一套简洁的 API，用于管理和操作缓存，可以快速为函数添加缓存功能，支持多种存储后端，并自带防缓存击穿、标签管理、Prometheus监控指标等高级功能。

## 亮点

1. ✅ **易于集成**：简单的 API 设计，和Golang设计理念一致，易于在现有项目中集成和使用。
2. ✅ **类型安全**：利用 Go 的泛型特性，自动推导，无需类型转换。
3. ✅ **防缓存击穿**：内置 singleflight 机制，有效防止缓存击穿问题。
4. ✅ **灵活的缓存策略**：支持自定义过期时间和基于标签的过期策略，满足各种缓存需求。
5. ✅ **可观测性**：内置Promtheus指标收集，便于监控和调试。
6. ✅ **多种缓存后端**：支持多种流行的缓存存储后端，灵活适应不同的应用场景。

## 安装

使用 go get 安装 go-cacheable：

```bash
go get github.com/diemus/go-cacheable
```

## 快速开始

### 初始化缓存管理器

首先，您需要初始化缓存管理器。go-cacheable 支持多种存储后端，这里我们以 Redis 和本地内存缓存为例，实际使用中您可以根据需要选择初始化哪些缓存管理器：

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
    // 初始化 Redis 客户端
    redisClient := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{":6379"},
    })
    redisStore := rediscluster.NewRedisCluster(redisClient)

    // 初始化本地缓存
    goCacheClient := cache.New(5*time.Minute, 10*time.Minute)
    goCacheStore := go_cache.NewGoCache(goCacheClient)

    // 创建缓存管理器
    RemoteCacheManager = cacheable.NewCacheManager(redisStore)
    LocalCacheManager = cacheable.NewCacheManager(goCacheStore)

    // 设置全局配置（可选）
    cacheable.SetDefaultKeyPrefix("myapp")
    cacheable.SetDefaultExpiration(5 * time.Minute)
}
```

### 基本使用

使用 `Get` 函数来包装一个函数，添加缓存能力：

改造前：
```go
func GetUser(ctx context.Context, id int) (User, error) {
    return fetchUserFromDatabase(id)
}
```

改造后，由于使用了泛型，因此返回值和原函数一致：
```go
func GetUser(ctx context.Context, id int) (User, error) {
    user, err, _ := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
        // 这里是当缓存未命中时获取用户信息的函数
        return fetchUserFromDatabase(id)
    })
    return user, err
}
```

不改造原函数，仅在调用时缓存，适合有时候需要缓存，有时候需要直接查库的情况：
```go
user, err, cached := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
    // 这里是当缓存未命中时获取用户信息的函数
    return GetUser(id)
})
```

Get函数可以灵活的指定使用远程缓存还是本地缓存：
```go
func GetUser(ctx context.Context, id int) (User, error) {
    user, err, _ := cacheable.Get(ctx, LocalCacheManager, "users", fmt.Sprintf("%d", id), func() (User, error) {
        // 这里是当缓存未命中时获取用户信息的函数
        return fetchUserFromDatabase(id)
    })
    return user, err
}
```

### 使用选项

go-cacheable 提供了多个选项来自定义缓存行为：

```go
user, err, _ := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), fetchUserFromDatabase,
    cacheable.WithExpiration(10 * time.Minute),
    cacheable.WithTags("user", fmt.Sprintf("teamId:%d", user.TeamId)),
    cacheable.WithDynamicTags(func() []string {
        // 这个函数只在设置缓存时被调用
        teamIDs, _ := getTeamIDsForUser("user1")
        tags := make([]string, len(teamIDs))
        for i, id := range teamIDs {
            tags[i] = fmt.Sprintf("teamId:%d", id)
        }
        return tags
    }),
)
```

### 标签的作用

标签用于给缓存定义元数据，便于批量删除。例如，如果缓存的 key 是 username，tag 可以是 teamId。当 team 发生变化时，可以删除所有与该 team 相关的用户缓存：

```go
// 设置缓存时添加 tag
value, err, _ := cacheable.Get(ctx, cacheManager, "users", username, getUserData,
    cacheable.WithTags(fmt.Sprintf("teamId:%d", userTeamID)),
)

// 当 team 发生变化时，删除相关缓存
err := cacheable.DeleteByTags(ctx, cacheManager, []string{fmt.Sprintf("teamId:%d", changedTeamID)})
```

### 动态标签

动态标签的目的是处理那些计算 tag 可能也是耗时操作的场景。例如，查找用户的 teamId 可能需要数据库查询。使用动态标签，这种计算只会在设置缓存时进行，而不会在每次获取缓存时重复计算：

```go
value, err, cached := cacheable.Get(ctx, cacheManager, "users", "user1", getUserData,
    cacheable.WithDynamicTags(func() []string {
        // 这个函数只在设置缓存时被调用
        teamIDs, _ := getTeamIDsForUser("user1")
        tags := make([]string, len(teamIDs))
        for i, id := range teamIDs {
            tags[i] = fmt.Sprintf("teamId:%d", id)
        }
        return tags
    }),
)
```


### 删除缓存

删除单个缓存项：

```go
err := cacheable.Delete(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id))
```

基于标签删除缓存，比如team变动时，清楚所有team成员的缓存：

```go
err := cacheable.DeleteByTags(ctx, RemoteCacheManager, []string{"teamId:123"})
```

### 多种缓存后端

go-cacheable 底层基于 [github.com/eko/gocache](https://github.com/eko/gocache)，支持多种缓存后端：

```go
// 安装所需的后端存储
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

// 使用示例（以 Redis 为例）
import "github.com/eko/gocache/store/redis/v4"

redisStore := redis.NewRedis(redisClient)
cacheManager := cacheable.NewCacheManager(redisStore)
```
## 指标收集

Go-Cacheable 内置了指标收集功能，可以轻松集成到您的监控系统中：

```go
// 指标会自动收集，您可以在您的指标系统中查看如下指标：
cacheable_cache_request_total{namespace="xxx"}
cacheable_cache_hit_total{namespace="xxx"}
```

## 配置

可以通过以下方法设置全局默认值：

```go
cacheable.SetDefaultKeyPrefix("cacheable")
cacheable.SetDefaultExpiration(30 * time.Minute)
cacheable.SetDefaultMetricsPrefix("cacheable")
```

## License

MIT

## 项目热度

[![Star History Chart](https://api.star-history.com/svg?repos=diemus/go-cacheable&type=Date)](https://star-history.com/#diemus/go-cacheable&Date)
