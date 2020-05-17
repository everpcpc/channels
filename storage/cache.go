package storage

type CacheBackend interface {
	Get(string) (string, error)
	Set(string, string) error
}

func Cached(cache CacheBackend, key string, fn func() (string, error)) (value string, err error) {
	value, err = cache.Get(key)
	if err == nil {
		return
	}

	value, err = fn()
	if err == nil {
		cache.Set(key, value)
	}
	return
}
