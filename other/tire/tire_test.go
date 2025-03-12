package tire

import (
	"fmt"
	"testing"
)

func TestTrie(t *testing.T) {
	trie := NewTrie()

	// 插入一些单词
	trie.Insert("你好")
	trie.Insert("你们好")
	trie.Insert("apple")
	trie.Insert("app")
	trie.Insert("banana")

	fmt.Println(trie.Search("apple"))  // true
	fmt.Println(trie.Search("app"))    // true
	fmt.Println(trie.Search("appl"))   // false
	fmt.Println(trie.Search("banana")) // true
	fmt.Println(trie.Search("你好"))     // true
}
