package internal

import (
	"maps"
	"sync"
)

type SharedCollection[T any] struct {
	objectsMap map[uint64]T
	nextId     uint64
	mapMutex   sync.Mutex
}

func NewSharedCollection[T any](capacity ...int) *SharedCollection[T] {
	var objectsMap map[uint64]T
	if len(capacity) > 0 {
		objectsMap = make(map[uint64]T, capacity[0])
	} else {
		objectsMap = make(map[uint64]T)
	}

	return &SharedCollection[T]{
		objectsMap: objectsMap,
		nextId:     1,
	}
}

func (sc *SharedCollection[T]) Add(obj T, customId ...uint64) uint64 {
	sc.mapMutex.Lock()
	defer sc.mapMutex.Unlock()

	id := sc.nextId
	if len(customId) > 0 {
		id = customId[0]
	}

	sc.objectsMap[id] = obj
	sc.nextId++
	return id
}

func (sc *SharedCollection[T]) Remove(id uint64) {
	sc.mapMutex.Lock()
	defer sc.mapMutex.Unlock()

	delete(sc.objectsMap, id)
}

func (sc *SharedCollection[T]) ForEach(callback func(uint64, T)) {
	sc.mapMutex.Lock()
	localCopy := make(map[uint64]T, len(sc.objectsMap))
	maps.Copy(localCopy, sc.objectsMap)
	sc.mapMutex.Unlock()

	for id, obj := range localCopy {
		callback(id, obj)
	}
}

func (sc *SharedCollection[T]) Get(id uint64) (T, bool) {
	sc.mapMutex.Lock()
	defer sc.mapMutex.Unlock()

	obj, ok := sc.objectsMap[id]
	return obj, ok
}

func (sc *SharedCollection[T]) Size() int {
	return len(sc.objectsMap)
}
