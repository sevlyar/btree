package btree

import (
	"log"
	"sync"
)

const debugIter = false

type Iterator struct {
	path []cursor
	n    *node
	i    int
	cur  Item
}

func NewIterator(tree *BTree) *Iterator {
	if tree.root == nil {
		return nil
	}
	i := iterPool.Get().(*Iterator)
	i.n = tree.root
	return i
}

var iterPool = sync.Pool{New: newIter}

func newIter() interface{} {
	return &Iterator{path: make([]cursor, 0, 16)}
}

type cursor struct {
	n *node
	i int
}

func (i *Iterator) Close() {
	i.path = i.path[:0]
	i.n = nil
	i.i = 0
	i.cur = nil
	iterPool.Put(i)
}

func (i *Iterator) Next(to Item) bool {
	for !i.endIsReach() {
		if !i.rightIsReach() {
			if i.hasChildren() {
				i.down(i.child())
				continue
			}
			i.cur = i.item()
			i.right()
			if to.Less(i.cur) {
				return false
			}
			return true
		}
		if i.i == len(i.n.items) {
			if i.hasChildren() {
				i.down(i.child())
				continue
			}
		}
		for {
			i.up()
			if i.endIsReach() {
				return false
			}
			if !i.rightIsReach() {
				i.cur = i.item()
				i.right()
				if to.Less(i.cur) {
					return false
				}
				return true
			}
		}
	}
	return false
}

func (i *Iterator) down(n *node) {
	if debugIter {
		log.Print("down: ", n.items)
	}
	i.path = append(i.path, cursor{i.n, i.i})
	i.n, i.i = n, 0
}

func (i *Iterator) up() {
	var top cursor
	if len(i.path) > 0 {
		top = i.path[len(i.path)-1]
		i.path = i.path[:len(i.path)-1]
	}
	i.n, i.i = top.n, top.i
	if debugIter {
		if i.n != nil {
			log.Print("up: ", i.i, i.n.items)
		} else {
			log.Print("top is reach")
		}
	}
}

func (i *Iterator) endIsReach() bool {
	return i.n == nil
}

func (i *Iterator) curs() *cursor {
	spot := &i.path[len(i.path)-1]
	if debugIter {
		log.Print("spot: ", spot.n.items, spot.n.children, spot.i)
	}
	return spot
}

func (i *Iterator) right() {
	if debugIter {
		log.Print("right: ", i.i)
	}
	i.i++
}

func (i *Iterator) item() Item {
	it := i.n.items[i.i]
	if debugIter {
		log.Print("item: ", it)
	}
	return it
}

func (i *Iterator) child() *node {
	return i.n.children[i.i]
}

func (i *Iterator) hasChildren() bool {
	return len(i.n.children) > 0
}

func (i *Iterator) rightIsReach() bool {
	return i.i >= len(i.n.items)
}

func (i *Iterator) SkipLess(than Item) {
	if debugIter {
		log.Print("skip less than: ", than)
		log.Print("down: ", i.n.items)
	}
	for !i.endIsReach() {
		// fall-down
		var found bool
		for idx, item := range i.n.items {
			if !item.Less(than) {
				i.i = idx
				found = true
				break
			}
		}
		if !found {
			i.i = len(i.n.items)
		}
		if i.hasChildren() {
			i.down(i.child())
			continue
		}
		return
	}
}

func (i *Iterator) Item() Item {
	return i.cur
}
