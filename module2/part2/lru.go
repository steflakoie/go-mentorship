package lru

import (
	"errors"
	"sync"
	"time"
)

// Элемент двусвязного списка
type node[K comparable, V any] struct {
	key      K
	value    V
	expireAt time.Time

	prev *node[K, V]
	next *node[K, V]
}

// Cache — потокобезопасный LRU-кэш с TTL-эвикцией.
// Нулевое значение не готово к использованию — создавай через New.
// После Close() кэш непригоден для использования (Get/Set вернут false/ничего не сделают).
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration

	items map[K]*node[K, V]

	head *node[K, V]
	tail *node[K, V]

	closed bool

	stopCh chan struct{}
	doneCh chan struct{}
}

// New создаёт кэш с заданной ёмкостью и TTL.
// capacity <= 0 — вернуть ошибку (или панику — реши и обоснуй).
// ttl <= 0 — TTL отключён (записи не истекают), только LRU-эвикция.
// Запускает фоновую горутину очистки (если ttl > 0).
func New[K comparable, V any](capacity int, ttl time.Duration) (*Cache[K, V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}

	cache := &Cache[K, V]{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[K]*node[K, V]),
	}

	return cache, nil
}

func (c *Cache[K, V]) removeNode(n *node[K, V]) {
	if n == nil {
		return
	}

	if n == c.head {
		c.head = n.next
	}

	if n == c.tail {
		c.tail = n.prev
	}

	if n.next != nil {
		n.next.prev = n.prev
	}
	if n.prev != nil {
		n.prev.next = n.next
	}

	n.next = nil
	n.prev = nil
}

func (c *Cache[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil

	if c.head == nil {
		n.next = nil
		c.head = n
		c.tail = n
	} else {
		c.head.prev = n
		n.next = c.head
		c.head = n
	}
}

func (c *Cache[K, V]) moveToFront(n *node[K, V]) {
	if n == nil || n == c.head {
		return
	}

	c.removeNode(n)
	c.pushFront(n)
}

// Get возвращает значение по ключу.
// Если ключ есть и не истёк — возвращает (value, true) и обновляет "recently used".
// Если ключ истёк — удаляет его и возвращает (zero, false).
// Если ключа нет — (zero, false).
func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zeroV V
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return zeroV, false
	}

	node, ok := c.items[key]
	if !ok {
		return zeroV, false
	}

	if c.ttl > 0 && !time.Now().Before(node.expireAt) {
		c.removeNode(node)
		delete(c.items, key)

		return zeroV, false
	}

	c.moveToFront(node)

	return node.value, true
}

// Set устанавливает значение с текущим TTL.
// Если capacity превышена — эвиктит least recently used.
// Если ключ уже есть — обновляет значение И expireAt.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	var expireAt time.Time
	if c.ttl > 0 {
		expireAt = time.Now().Add(c.ttl)
	}

	n, ok := c.items[key]
	if ok {
		n.value = value
		n.expireAt = expireAt
		c.moveToFront(n)
		return
	}

	if len(c.items) >= c.capacity {
		tailKey := c.tail.key
		c.removeNode(c.tail)
		delete(c.items, tailKey)
	}

	newNode := &node[K, V]{
		key:      key,
		value:    value,
		expireAt: expireAt,
	}

	c.pushFront(newNode)
	c.items[key] = newNode
}

// Delete удаляет ключ. No-op если ключа нет.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	n, ok := c.items[key]
	if !ok {
		return
	}

	c.removeNode(n)
	delete(c.items, key)
}

// Len возвращает текущее число записей (включая истёкшие, ещё не убранные фоновой очисткой).
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

// Close останавливает фоновую горутину очистки.
// Идемпотентен — повторный вызов безопасен.
func (c *Cache[K, V]) Close() {
	c.mu.Lock()

	if !c.closed {
		c.closed = true
		close(c.stopCh)
	}
	c.mu.Unlock()

	<-c.doneCh
}

func (c *Cache[K, V]) cleanupLoop() {
	defer close(c.doneCh)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache[K, V]) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl <= 0 {
		return
	}
	now := time.Now()

	for h := c.head; h != nil; {
		next := h.next
		if !now.Before(h.expireAt) {
			c.removeNode(h)
			delete(c.items, h.key)

		}
		h = next
	}
}
