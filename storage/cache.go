package storage

type CacheBackend interface {
	Get(string) (string, error)
	Set(string, string) error
}
