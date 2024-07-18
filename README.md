# Cacheable

Cacheable 是一个强大而灵活的 Go 语言缓存库，旨在简化缓存操作并提高应用程序性能。它提供了一套简洁的 API，支持多种缓存存储后端，并具有丰富的功能和优化。

## 功能特性

- 支持多种缓存存储后端（如 Redis、本地内存缓存等）
- 泛型支持，适用于各种数据类型
- 防缓存击穿（使用 singleflight）
- 灵活的缓存选项（过期时间、标签等）
- 动态标签支持
- 内置指标收集
- 简洁易用的 API

## 亮点

1. **类型安全**：利用 Go 的泛型特性，确保类型安全，减少运行时错误。
2. **防缓存击穿**：内置 singleflight 机制，有效防止缓存击穿问题。
3. **灵活的缓存策略**：支持自定义过期时间和标签，满足各种缓存需求。
4. **动态标签**：支持在缓存时动态计算标签，适用于复杂的缓存场景。
5. **性能优化**：针对高并发场景进行优化，提供卓越的性能。
6. **易于集成**：简单的 API 设计，易于在现有项目中集成和使用。
7. **可观测性**：内置指标收集，便于监控和调试。

## 安装

使用 go get 安装 Cacheable：

```bash
go get github.com/yourusername/cacheable
```

## 快速开始

以下是一个简单的示例，展示如何使用 Cacheable：

```go
package main

import (
"context"
"fmt"
"time"

"github.com/yourusername/cacheable"
"github.com/eko/gocache/lib/v4/store"
"github.com/eko/gocache/store/redis/v4"
)

func main() {
// 初始化 Redis 存储
redisStore := redis.NewRedis(redisClient)

// 创建缓存管理器
cacheManager := cacheable.NewCacheManager(redisStore)

// 使用缓存
ctx := context.Background()
value, err, cached := cacheable.Get(ctx, cacheManager, "users", "user1", func() (string, error) {
// 这个函数只有在缓存未命中时才会被调用
return "John Doe", nil
})

if err != nil {
fmt.Println("Error:", err)
return
}

if cached {
fmt.Println("Value from cache:", value)
} else {
fmt.Println("Value from function:", value)
}
}
```

## 高级用法

### 自定义选项

Cacheable 支持多种自定义选项，如设置过期时间和标签：

```go
value, err, cached := cacheable.Get(ctx, cacheManager, "users", "user1", getUserData,
cacheable.WithExpiration(10*time.Minute),
cacheable.WithTags("user", "profile"),
)
```

### 动态标签

对于需要在运行时计算标签的场景，可以使用动态标签功能：

```go
value, err, cached := cacheable.Get(ctx, cacheManager, "users", "user1", getUserData,
cacheable.WithDynamicTags(func() []string {
return []string{fmt.Sprintf("org:%d", getCurrentOrgID())}
}),
)
```

### 删除缓存

可以通过键或标签删除缓存：

```go
// 通过键删除
err := cacheable.Delete(ctx, cacheManager, "users", "user1")

// 通过标签删除
err := cacheable.DeleteByTags(ctx, cacheManager, []string{"org:123"})
```

### 使用本地缓存

Cacheable 支持多种缓存后端，包括本地内存缓存：

```go
import (
"github.com/eko/gocache/store/go_cache/v4"
gocache "github.com/patrickmn/go-cache"
)

func main() {
gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
localStore := go_cache.NewGoCache(gocacheClient)
localCacheManager := cacheable.NewCacheManager(localStore)

// 使用本地缓存管理器...
}
```

### 多级缓存

可以轻松地实现多级缓存策略：

```go
func GetUserData(ctx context.Context, userID string) (User, error) {
// 尝试从本地缓存获取
value, err, cached := cacheable.Get(ctx, localCacheManager, "users", userID, func() (User, error) {
// 如果本地缓存未命中，尝试从远程缓存获取
return cacheable.Get(ctx, remoteCacheManager, "users", userID, func() (User, error) {
// 如果远程缓存也未命中，从数据库获取
return getUserFromDatabase(userID)
})
})

return value, err
}
```

## 性能优化

Cacheable 使用 singleflight 来防止缓存击穿，这在高并发场景下特别有用：

```go
// 这个调用会自动使用 singleflight 来防止缓存击穿
value, err, cached := cacheable.Get(ctx, cacheManager, "high_concurrency", "key", expensiveOperation)
```

## 指标收集

Cacheable 内置了指标收集功能，可以轻松集成到您的监控系统中：

```go
// 指标会自动收集，您可以在您的指标系统中查看如下指标：
// cacheable_request_total
// cacheable_hit_total
```

## 配置

可以通过以下方法设置全局默认值：

```go
cacheable.SetDefaultKeyPrefix("myapp")
cacheable.SetDefaultExpiration(30 * time.Minute)
cacheable.SetDefaultMetricsPrefix("myapp_cache")
```

## 贡献

欢迎贡献代码、报告问题或提出改进建议。请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解更多信息。

## 许可证

本项目采用 MIT 许可证。详情请见 [LICENSE](LICENSE) 文件。