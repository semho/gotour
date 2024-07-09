package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	key        string
	value      any
	expiration time.Time
	element    *list.Element
}

type Cache struct {
	capacity    int
	items       map[string]any
	list        *list.List
	mu          *sync.RWMutex
	stopCleaner chan struct{}
}

type CacheOption func(*Cache)

func WithCapacity(capacity int) CacheOption {
	return func(c *Cache) {
		c.capacity = capacity
		c.list = list.New()
	}
}

func NewCache(options ...CacheOption) *Cache {
	c := &Cache{items: make(map[string]any), mu: &sync.RWMutex{}}
	for _, option := range options {
		option(c)
	}
	if c.capacity == 0 { // TTL кеш
		c.stopCleaner = make(chan struct{})
		go c.cleanupLoop()
	}
	return c
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleaner:
			return
		}
	}
}

// TODO: переделать блокировки. Сначала все читаем через RLock. Ищем у кого закончилось время. Элементы добавляем
// TODO: в отдельный слайс. После этого отдельно проходимся по этому слайсу, удаляем все элементы, используя Lock.
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, val := range c.items {
		item := val.(*CacheItem)
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

func (c *Cache) Close() {
	if c.stopCleaner != nil {
		close(c.stopCleaner)
	}
}

func (c *Cache) muLoad(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	if !ok {
		return nil, ok
	}

	return val, ok
}

func (c *Cache) muMoveToFront(item *CacheItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list.MoveToFront(item.element)
}

func (c *Cache) Get(key string) (any, bool) {
	val, ok := c.muLoad(key)
	if !ok {
		return nil, ok
	}
	item := val.(*CacheItem)
	if time.Now().After(item.expiration) && !c.isLRU() {
		c.Delete(item)
		return nil, false
	}

	if c.isLRU() {
		c.muMoveToFront(item)
	}

	return item.value, ok
}

func (c *Cache) isLRU() bool {
	if c.list != nil {
		return true
	}

	return false
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	newItem := &CacheItem{key: key, value: value, expiration: expiration}

	if c.isLRU() {
		if val, ok := c.muLoad(key); ok {
			item := val.(*CacheItem)
			c.mu.Lock()
			item.expiration = expiration
			item.value = value
			c.mu.Unlock()
			return
		}

		if c.capacity == c.list.Len() {
			c.mu.RLock()
			oldElement := c.list.Back()
			oldItem := oldElement.Value.(*CacheItem)
			c.mu.RUnlock()
			if oldItem != nil {
				c.Delete(oldItem)
			}
		}
		c.mu.Lock()
		newElement := c.list.PushFront(newItem)
		newItem.element = newElement
		c.mu.Unlock()
	}
	c.mu.Lock()
	c.items[key] = newItem
	c.mu.Unlock()
}

func (c *Cache) Delete(item *CacheItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, item.key)
	if c.isLRU() {
		c.list.Remove(item.element)
	}
}

func main() {
	//LRU кэш
	lruCache := NewCache(WithCapacity(2))

	lruCache.Set("key1", "value1", 0) // TTL не используется для LRU
	lruCache.Set("key2", "value2", 0)
	lruCache.Set("key3", "value3", 3*time.Minute)

	valueLru, okLru := lruCache.Get("key2")
	if okLru {
		fmt.Printf("LRU Cache - Key: key2, Value: %v\n", valueLru)
	}

	lruCache.Set("key4", "value4", 3*time.Minute)

	valueLru2, okLru2 := lruCache.Get("key2")
	if okLru2 {
		fmt.Printf("LRU Cache - Key: key2, Value: %v\n", valueLru2)
	}

	// TTL кеш (без ограничения емкости)
	ttlCache := NewCache()
	defer ttlCache.Close()

	ttlCache.Set("key1", "value1", 3*time.Second)
	ttlCache.Set("key2", "value2", 2*time.Minute)

	value, ok := ttlCache.Get("key1")
	if ok {
		fmt.Printf("TTL Cache - Key: key1, Value: %v\n", value)
	}
	fmt.Println(ttlCache)
	//time.Sleep(5 * time.Second)
	//value2, ok2 := ttlCache.Get("key1")
	//if ok2 {
	//	fmt.Printf("TTL Cache - Key: key1, Value: %v\n", value2)
	//}
	//fmt.Println(ttlCache)
}
