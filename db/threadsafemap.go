package crdb

import "sync"

type ThreadSafeMap struct {
    sync.RWMutex
    dict map[interface{}]interface{}
}

func NewThreadSafeMap() ThreadSafeMap {
    return ThreadSafeMap{dict: make(map[interface{}]interface{})}
}

func (d ThreadSafeMap) Insert(k, v interface{}) bool {
    d.Lock()
    defer d.Unlock()
    _, ok := d.dict[k]
    if !ok { d.dict[k] = v }
    return !ok
}

func (d ThreadSafeMap) Remove(k interface{}) bool {
    d.Lock()
    defer d.Unlock()
    _, ok := d.dict[k]
    delete(d.dict, k)
    return ok
}

func (d ThreadSafeMap) GetValue(k interface{}) interface{} {
    d.RLock()
    defer d.RUnlock()
    v, _ := d.dict[k]
    return v
}

func (d ThreadSafeMap) Keys() []interface{} {
    d.RLock()
    defer d.RUnlock()
    results := make([]interface{}, 0)
    for k, _ := range d.dict { results = append(results, k) }
    return results
}

func (d ThreadSafeMap) Contains(k interface{}) bool {
    d.RLock()
    defer d.RUnlock()
    _, ok := d.dict[k]
    return ok
}

