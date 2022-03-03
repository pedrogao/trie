package trie

import (
	"strings"
)

// FuzzyTrie fuzzy search、put、delete trie
type FuzzyTrie struct {
	segmenter StringSegmenter
	value     interface{}
	children  map[string]*FuzzyTrie
}

// FuzzyTrieConfig for building a path trie with different segmenter
type FuzzyTrieConfig struct {
	Segmenter StringSegmenter
}

// NewFuzzyTrie return default *FuzzyTrie
func NewFuzzyTrie() *FuzzyTrie {
	return &FuzzyTrie{
		segmenter: PathSegmenter,
	}
}

// NewFuzzyTrieWithConfig allocates and returns a new *FuzzyTrie with the given *FuzzyTrieConfig
func NewFuzzyTrieWithConfig(config *FuzzyTrieConfig) *FuzzyTrie {
	segmenter := PathSegmenter
	if config != nil && config.Segmenter != nil {
		segmenter = config.Segmenter
	}

	return &FuzzyTrie{
		segmenter: segmenter,
	}
}

// Get 获取数据，精准匹配
func (f *FuzzyTrie) Get(key string) interface{} {
	node := f // 当前节点
	// 从 key 中寻找 segment，比如 /root/usr 找到的第一个 segment 是 /root
	// 然后依次向后寻找
	for part, i := f.segmenter(key, 0); part != ""; part, i = f.segmenter(key, i) {
		node = node.children[part] // 子节点
		if node == nil {
			return nil
		}
	}
	return node.value
}

// Put 插入数据，精准匹配
// @return false is put fail
func (f *FuzzyTrie) Put(key string, value interface{}) bool {
	node := f
	for part, i := f.segmenter(key, 0); part != ""; part, i = f.segmenter(key, i) {
		if strings.HasSuffix(part, "*") {
			// can't put with * suffix
			return false
		}
		child := node.children[part]
		if child == nil {
			if node.children == nil {
				node.children = map[string]*FuzzyTrie{}
			}
			child = f.newFuzzyTrie()
			node.children[part] = child
		}
		node = child
	}
	node.value = value
	return true
}

// Delete 删除数据，可模糊删除
// @return true if value is found by the given key
func (f *FuzzyTrie) Delete(key string) bool {
	nodes := make([]*fuzzyNode, 0)
	node := f
	for part, i := f.segmenter(key, 0); part != ""; part, i = f.segmenter(key, i) {
		// 加入当前路过的node
		nodes = append(nodes, &fuzzyNode{
			node: node,
			part: part,
		})
		node = node.children[part] // 子节点
		star := strings.HasSuffix(part, "*")
		if node == nil {
			if !star {
				return false
			}
			// handle star
			if strings.TrimSpace(part) == "/*" {
				// 删除 node 的所有孩子节点，node 本身无改变
				parent := nodes[len(nodes)-1].node
				parent.children = nil
				return true
			} else {
				prefix := strings.TrimSuffix(part, "*")
				// 删除 node 匹配 part 的孩子节点
				parent := nodes[len(nodes)-1].node
				for k := range parent.children {
					if strings.HasPrefix(k, prefix) {
						delete(parent.children, k)
					}
				}
				return true
			}
		}
	}

	// delete value
	node.value = nil
	if node.isLeaf() {
		// 从父节点中删除 key
		for i := len(nodes) - 1; i >= 0; i-- {
			parent := nodes[i].node
			part := nodes[i].part
			delete(parent.children, part)
			if !parent.isLeaf() {
				break
			}
			// parent 已经是叶子节点了
			parent.children = nil
			if parent.value != nil {
				// 父节点有值，直接 break
				break
			}
		}
	} else {
		// 非叶子节点，孩子也被删掉
		// 值虽然被删除了，但是节点还在
		node.children = nil
	}
	return true
}

func (f *FuzzyTrie) Walk(walker WalkFunc) error {
	return f.walk("", walker)
}

// WalkPath 必须满足模糊匹配
// 如：/usr/local/bin; /usr/local/env 都可通过 /usr/local 来搜索
// /usr/local/bin; /usr/local/bit 都可通过 /usr/local/b* 来搜索，而
// /usr/local/env 不行
func (f *FuzzyTrie) WalkPath(key string, walker WalkFunc) error {
	// 先遍历根节点
	if f.value != nil {
		if err := walker("", f.value); err != nil {
			return err
		}
	}
	node := f
	parent := node
	// 依次遍历子节点
	for part, i := f.segmenter(key, 0); part != ""; part, i = f.segmenter(key, i) {
		parent = node
		node = node.children[part]
		var k string
		if i == -1 { // -1 表示没有分割成功，所有 k = key
			k = key
		} else {
			k = key[0:i]
		}
		if node == nil {
			if !strings.HasSuffix(part, "*") {
				return nil
			}
			// handle star
			if part == "/*" {
				// walk parent
				for k1, v := range parent.children {
					if err := walker(strings.TrimSuffix(k, part)+k1, v.value); err != nil {
						return err
					}
				}
			} else {
				prefix := strings.TrimSuffix(part, "*")
				// 删除 node 匹配 part 的孩子节点
				for k1, v := range parent.children {
					if !strings.HasPrefix(k1, prefix) {
						continue
					}
					if err := walker(strings.TrimSuffix(k, part)+k1, v.value); err != nil {
						return err
					}
				}
			}
		}
		if node != nil && node.value != nil {
			if err := walker(k, node.value); err != nil {
				return err
			}
		}
		if i == -1 {
			break
		}
	}
	return nil
}

// walk iterates all children
func (f *FuzzyTrie) walk(key string, walker WalkFunc) error {
	if f.value != nil {
		if err := walker(key, f.value); err != nil {
			return err
		}
	}
	for part, child := range f.children {
		if err := child.walk(key+part, walker); err != nil {
			return err
		}
	}
	return nil
}

func (f *FuzzyTrie) newFuzzyTrie() *FuzzyTrie {
	return &FuzzyTrie{
		segmenter: f.segmenter,
	}
}

// isLeaf end of trie
func (f *FuzzyTrie) isLeaf() bool {
	return len(f.children) == 0
}

// fuzzyNode represent a fuzzy node
type fuzzyNode struct {
	node *FuzzyTrie // node instance
	part string     // node sub path
}
