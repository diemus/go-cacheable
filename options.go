package cacheable

import "time"

type Option func(o *Options)

type Options struct {
	Expiration time.Duration
	Tags       []string
}

func applyOptions(opts ...Option) *Options {
	o := &Options{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithTags(tags ...string) Option {
	return func(o *Options) {
		o.Tags = append(o.Tags, tags...)
	}
}

// WithDynamicTags 动态添加tags，适合计算tag需要做耗时操作的场景，仅在set缓存时进行tag计算
func WithDynamicTags(fn func() []string) Option {
	return func(o *Options) {
		o.Tags = append(o.Tags, fn()...)
	}
}

func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.Expiration = expiration
	}
}
