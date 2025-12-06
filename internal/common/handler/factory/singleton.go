package factory

import "sync"

type Supplier func(string) any

type Singleton struct {
	mu       sync.Mutex
	cache    map[string]any
	supplier Supplier
}

func NewSingleton(supplier Supplier) *Singleton {
	return &Singleton{
		cache:    make(map[string]any),
		supplier: supplier,
	}
}

// Get 获取 key 对应的单例实例，若不存在则通过 supplier 创建并缓存
func (s *Singleton) Get(key string) any {
	if value, hit := s.cache[key]; hit {
		return value
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if value, hit := s.cache[key]; hit {
		return value
	}

	s.cache[key] = s.supplier(key)
	return s.cache[key]
}
