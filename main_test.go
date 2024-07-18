package cacheable

import (
	"github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"os"
	"testing"
	"time"
)

var MockCacheManager *CacheManager

func TestMain(m *testing.M) {

	// 初始化
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := go_cache.NewGoCache(gocacheClient)
	MockCacheManager = NewCacheManager(gocacheStore)

	// 运行测试
	code := m.Run()

	// 在这里进行清理工作（如果需要）

	// 退出测试
	os.Exit(code)
}
