package cacheable

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/eko/gocache/lib/v4/store"
	"golang.org/x/sync/singleflight"
	"time"
)

var defaultKeyPrefix = "cacheable"
var defaultExpiration = 60 * time.Minute
var defaultMetricsPrefix = "cacheable"

type CacheManager struct {
	sg    singleflight.Group
	cache store.StoreInterface
}

func NewCacheManager(store store.StoreInterface) *CacheManager {
	return &CacheManager{
		sg:    singleflight.Group{},
		cache: store,
	}
}

func (i *CacheManager) Get(ctx context.Context, namespace string, key string, fn func() ([]byte, error), opts ...Option) (value []byte, err error, cached bool) {
	CacheRequestTotal.WithLabelValues(namespace).Inc()
	// 拼接namespace和key作为缓存的key
	key = defaultKeyPrefix + ":" + namespace + ":" + key
	data, err := i.cache.Get(ctx, key)
	if err != nil && !errors.Is(err, store.NotFound{}) {
		//非缓存不存在错误，直接返回
		return nil, err, false
	} else if err == nil {
		//缓存存在，直接返回
		CacheHitTotal.WithLabelValues(namespace).Inc()
		//这里有个bug，redis取出的是string, go-cache取出的是[]byte，需要做类型转换
		switch data.(type) {
		case []byte:
			return data.([]byte), nil, true
		case string:
			return []byte(data.(string)), nil, true
		default:
			return nil, errors.New("unsupported data type"), false
		}
	}

	//缓存不存在，调用fn获取数据，使用single flight防止缓存击穿
	result, fnErr, _ := i.sg.Do(key, func() (interface{}, error) {
		d, err := fn()
		if err != nil {
			return nil, err
		}
		return d, nil
	})

	if fnErr != nil {
		return nil, fnErr, false
	}

	data, ok := result.([]byte)
	if !ok {
		return nil, errors.New("result type error"), false
	}

	//将自定义的Option转换为store.Option
	options := applyOptions(opts...)
	var setOptions []store.Option
	if options.Expiration > 0 {
		setOptions = append(setOptions, store.WithExpiration(options.Expiration))
	} else {
		setOptions = append(setOptions, store.WithExpiration(defaultExpiration))
	}
	if len(options.Tags) > 0 {
		setOptions = append(setOptions, store.WithTags(options.Tags))
	}

	err = i.cache.Set(ctx, key, data, setOptions...)
	if err != nil {
		return nil, err, false
	}

	return data.([]byte), fnErr, false
}

func (i *CacheManager) Delete(ctx context.Context, namespace string, key string) error {
	// 拼接namespace和key作为缓存的key
	key = defaultKeyPrefix + ":" + namespace + ":" + key
	return i.cache.Delete(ctx, key)
}

func (i *CacheManager) DeleteByTags(ctx context.Context, tags []string) error {
	return i.cache.Invalidate(ctx, store.WithInvalidateTags(tags))
}

// Get 尝试从缓存中获取值，如果没有则调用 fn 获取并缓存，这里使用了泛型来支持不同类型的返回值，同时支持options的方式给缓存添加tag和有效期
func Get[T any](ctx context.Context, cacheManager *CacheManager, namespace string, key string, fn func() (T, error), opts ...Option) (value T, err error, cached bool) {
	data, err, cached := cacheManager.Get(ctx, namespace, key, func() ([]byte, error) {
		v, e := fn()
		if e != nil {
			return nil, e
		}
		return json.Marshal(v)
	}, opts...)
	if err != nil {
		return value, err, cached
	}

	err = json.Unmarshal(data, &value)
	if err != nil {
		return value, err, cached
	}
	return value, err, cached
}

func Delete(ctx context.Context, cacheManager *CacheManager, namespace string, key string) error {
	return cacheManager.Delete(ctx, namespace, key)
}

func DeleteByTags(ctx context.Context, cacheManager *CacheManager, tags []string) error {
	return cacheManager.DeleteByTags(ctx, tags)
}

func SetDefaultKeyPrefix(prefix string) {
	defaultKeyPrefix = prefix
}

func SetDefaultExpiration(expiration time.Duration) {
	defaultExpiration = expiration
}

func SetDefaultMetricsPrefix(prefix string) {
	defaultMetricsPrefix = prefix
}
