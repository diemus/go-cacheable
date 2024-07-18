package main

import (
	"context"
	"fmt"
	"github.com/diemus/go-cacheable"
	"github.com/eko/gocache/lib/v4/store/go_cache"
	"github.com/eko/gocache/lib/v4/store/rediscluster"
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

func GetUser(ctx context.Context, id int) (User, error) {
	user, err, _ := cacheable.Get(ctx, RemoteCacheManager, "users", fmt.Sprintf("%d", id), fetchUserFromDatabase,
		cacheable.WithExpiration(10*time.Minute),
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
	return user, err
}

func main() {
	InitCacheManager()
}
