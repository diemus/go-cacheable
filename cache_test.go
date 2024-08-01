package cacheable

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := context.Background()

	t.Run("缓存命中", func(t *testing.T) {
		key := "hit"
		expected := "cached value"
		_, err, _ := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return expected, nil
		})
		assert.NoError(t, err)

		start := time.Now()
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.True(t, cached)
		assert.Equal(t, expected, value)
		assert.Less(t, duration, 100*time.Millisecond) // 假设缓存命中应该很快
	})

	t.Run("缓存未命中", func(t *testing.T) {
		key := "miss"
		expected := "new value"
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			time.Sleep(50 * time.Millisecond) // 模拟耗时操作
			return expected, nil
		})

		assert.NoError(t, err)
		assert.False(t, cached)
		assert.Equal(t, expected, value)
	})

	t.Run("fn返回错误", func(t *testing.T) {
		key := "error"
		expectedErr := errors.New("fn error")
		_, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "", expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, cached)
	})

	t.Run("fn返回空数据", func(t *testing.T) {
		key := "empty"
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "", nil
		})

		assert.NoError(t, err)
		assert.False(t, cached)
		assert.Empty(t, value)
	})

	t.Run("fn返回不同数据类型", func(t *testing.T) {
		key := "different_type"
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (int, error) {
			return 123, nil
		})

		assert.NoError(t, err)
		assert.False(t, cached)
		assert.Equal(t, 123, value)
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("删除存在的缓存", func(t *testing.T) {
		key := "existing"
		_, _, _ = Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "value", nil
		})

		err := Delete(ctx, MockCacheManager, namespace, key)
		assert.NoError(t, err)

		// 验证缓存已被删除
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.NoError(t, err)
		assert.False(t, cached)
		assert.Equal(t, "new value", value)
	})

	t.Run("删除不存在的缓存", func(t *testing.T) {
		key := "non_existing"
		err := Delete(ctx, MockCacheManager, namespace, key)
		assert.NoError(t, err)
	})

	t.Run("删除缓存的幂等性", func(t *testing.T) {
		key := "idempotent"
		_, _, _ = Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "value", nil
		})

		err1 := Delete(ctx, MockCacheManager, namespace, key)
		err2 := Delete(ctx, MockCacheManager, namespace, key)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestDeleteByTags(t *testing.T) {
	ctx := context.Background()

	t.Run("根据标签删除缓存", func(t *testing.T) {
		tag := "tag1"
		key1, key2 := "key1", "key2"

		_, _, _ = Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "value1", nil
		}, WithTags(tag))
		_, _, _ = Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "value2", nil
		}, WithTags(tag))

		err := DeleteByTags(ctx, MockCacheManager, []string{tag})
		assert.NoError(t, err)

		// 验证缓存已被删除
		_, _, cached1 := Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "new value1", nil
		})
		_, _, cached2 := Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "new value2", nil
		})

		assert.False(t, cached1)
		assert.False(t, cached2)

		//需要删除缓存，如果是redis，里面还留着旧的key会导致单测异常
		Delete(ctx, MockCacheManager, namespace, key1)
		Delete(ctx, MockCacheManager, namespace, key2)
	})

	t.Run("根据不存在的标签删除缓存", func(t *testing.T) {
		err := DeleteByTags(ctx, MockCacheManager, []string{"non_existing_tag"})
		assert.NoError(t, err)
	})

	t.Run("根据多个标签删除缓存", func(t *testing.T) {
		tag1, tag2 := "tag1", "tag2"
		key1, key2, key3 := "key1", "key2", "key3"

		_, _, _ = Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "value1", nil
		}, WithTags(tag1))
		_, _, _ = Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "value2", nil
		}, WithTags(tag2))
		_, _, _ = Get(ctx, MockCacheManager, namespace, key3, func() (string, error) {
			return "value3", nil
		}, WithTags(tag1, tag2))

		err := DeleteByTags(ctx, MockCacheManager, []string{tag1, tag2})
		assert.NoError(t, err)

		// 验证缓存已被删除
		_, _, cached1 := Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "new value1", nil
		})
		_, _, cached2 := Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "new value2", nil
		})
		_, _, cached3 := Get(ctx, MockCacheManager, namespace, key3, func() (string, error) {
			return "new value3", nil
		})

		assert.False(t, cached1)
		assert.False(t, cached2)
		assert.False(t, cached3)

		//需要删除缓存，如果是redis，里面还留着旧的key会导致单测异常
		Delete(ctx, MockCacheManager, namespace, key1)
		Delete(ctx, MockCacheManager, namespace, key2)
		Delete(ctx, MockCacheManager, namespace, key3)
	})
}

func TestGetWithOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("使用过期时间选项", func(t *testing.T) {
		key := "expiration"
		expected := "value"
		expiration := 100 * time.Millisecond

		_, err, _ := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return expected, nil
		}, WithExpiration(expiration))
		assert.NoError(t, err)

		// 立即获取，应该命中缓存
		value, err, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.NoError(t, err)
		assert.True(t, cached)
		assert.Equal(t, expected, value)

		// 等待过期
		time.Sleep(expiration + 10*time.Millisecond)

		// 再次获取，应该未命中缓存
		value, err, cached = Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.NoError(t, err)
		assert.False(t, cached)
		assert.Equal(t, "new value", value)
	})

	t.Run("使用标签选项", func(t *testing.T) {
		key1, key2 := "tag_test1", "tag_test2"
		tag := "test_tag"

		_, err, _ := Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "value1", nil
		}, WithTags(tag))
		assert.NoError(t, err)

		_, err, _ = Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "value2", nil
		}, WithTags(tag))
		assert.NoError(t, err)

		// 使用标签删除缓存
		err = DeleteByTags(ctx, MockCacheManager, []string{tag})
		assert.NoError(t, err)

		// 验证缓存已被删除
		_, _, cached1 := Get(ctx, MockCacheManager, namespace, key1, func() (string, error) {
			return "new value1", nil
		})
		_, _, cached2 := Get(ctx, MockCacheManager, namespace, key2, func() (string, error) {
			return "new value2", nil
		})

		assert.False(t, cached1)
		assert.False(t, cached2)
	})

	t.Run("使用多个标签", func(t *testing.T) {
		key := "multi_tag"
		tag1, tag2 := "tag1", "tag2"

		_, err, _ := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "value", nil
		}, WithTags(tag1, tag2))
		assert.NoError(t, err)

		// 使用第一个标签删除缓存
		err = DeleteByTags(ctx, MockCacheManager, []string{tag1})
		assert.NoError(t, err)

		// 验证缓存已被删除
		_, _, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.False(t, cached)
	})

	t.Run("同时使用过期时间和标签选项", func(t *testing.T) {
		key := "expiration_and_tag"
		tag := "test_tag"
		expiration := 100 * time.Millisecond

		_, err, _ := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "value", nil
		}, WithExpiration(expiration), WithTags(tag))
		assert.NoError(t, err)

		// 立即获取，应该命中缓存
		_, _, cached := Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.True(t, cached)

		// 使用标签删除缓存
		err = DeleteByTags(ctx, MockCacheManager, []string{tag})
		assert.NoError(t, err)

		// 验证缓存已被删除
		_, _, cached = Get(ctx, MockCacheManager, namespace, key, func() (string, error) {
			return "new value", nil
		})
		assert.False(t, cached)
	})
}
