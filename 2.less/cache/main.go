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
	expiration int64
	element    *list.Element
}

type LRUCache struct {
	capacity int
	items    sync.Map
	list     *list.List
	mu       sync.RWMutex
}

func NewCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     list.New(),
	}
}

func (c *LRUCache) Len() {
	count := 0
	c.items.Range(
		func(_, _ interface{}) bool {
			count++
			return true
		},
	)
	fmt.Printf("количество элементов в карте: %d\n", count)
	fmt.Printf("количество элементов в списке: %d\n", c.list.Len())
}

func (c *LRUCache) GetList() {
	for iterator := c.list.Front(); iterator != nil; iterator = iterator.Next() {
		fmt.Printf("key: %s, value: %v\n", iterator.Value.(*CacheItem).key, iterator.Value.(*CacheItem).value)
	}
}

func (c *LRUCache) getItem(key string) (*CacheItem, bool) {
	val, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}

	item := val.(*CacheItem)
	if c.timeOver(item) {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.list.MoveToFront(item.element)

	return item, true
}

func (c *LRUCache) timeOver(item *CacheItem) bool {
	if item.expiration <= time.Now().Unix() {
		c.Delete(item)
		return true
	}
	return false
}

func (c *LRUCache) Delete(item *CacheItem) {
	c.items.Delete(item.key)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list.Remove(item.element)
}

func (c *LRUCache) Get(key string) any {
	if item, ok := c.getItem(key); ok {
		return item.value
	}

	return nil
}

func (c *LRUCache) Set(key string, value any, ttl time.Duration) {
	expiration := time.Now().Add(ttl).Unix()
	//уже существует и время не вышло
	if val, ok := c.getItem(key); ok {
		val.value = value
		val.expiration = expiration
		return
	}

	if c.capacity == c.list.Len() {
		//удалить последний элемент
		c.mu.RLock()
		oldElement := c.list.Back()
		oldItem := oldElement.Value.(*CacheItem)
		c.mu.RUnlock()
		if oldItem != nil {
			c.Delete(oldItem)
		}
	}

	item := &CacheItem{
		key:        key,
		value:      value,
		expiration: expiration,
	}

	c.mu.Lock()
	element := c.list.PushFront(item)
	c.mu.Unlock()
	//добавляем элемент из списка в структуру
	item.element = element
	c.items.Store(key, item)
}

func main() {
	cache := NewCache(3)
	cache.Set("key1", "value1", 5*time.Second)
	cache.Set("key2", "value2", 5*time.Second)
	cache.Set("key3", "value3", 3*time.Second)
	cache.Set("key4", "value4", 4*time.Second)
	cache.Len()
	cache.GetList()
	time.Sleep(3 * time.Second)
	fmt.Println(cache.Get("key1"))
	fmt.Println(cache.Get("key2"))
	fmt.Println(cache.Get("key3"))
	fmt.Println(cache.Get("key4"))
	cache.Set("key4", "value4", 4*time.Second)
	fmt.Println(cache.Get("key2"))
	cache.Len()
	cache.GetList()
	fmt.Println(cache.Get("key5"))
}
