// // Copyright (c) 2026 Pranav R S All rights reserved.
// // Use of this source code is governed by a BSD-style
// // license that can be found in the LICENSE file.

package tlru_test

import (
	"fmt"

	"github.com/justpranavrs/tlru"
)

// Member is the type of the value stored in the cache.
type Member struct {
	Name  string
	Email string
}

// ExampleCache shows a small example of how to initialize a LRU instance and
// do basic operations like Put, Size, Contains and Capacity.
func ExampleCache() {
	cache, err := tlru.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{ // insert in user data with user id 1
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})

	fmt.Println(cache.Size()) // gets the current size of the cache
	fmt.Println(cache.Contains(2))

	fmt.Println(cache.Contains(1)) // reports whether key 1 is present in the cache
	fmt.Println(cache.Capacity())

	cache.Flush()
	fmt.Println(cache.Size())
	fmt.Println(cache.Capacity()) // capacity allocated for the cache

	// Output:
	// 1
	// false
	// true
	// 256
	// 0
	// 256
}

// ExampleLRU_Capacity shows an example of how Capacity works.
func ExampleLRU_Capacity() {
	cache, err := tlru.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	fmt.Println(cache.Capacity())

	// Output:
	// 256
}

// ExampleLRU_Contains shows an example of how Contains works.
func ExampleLRU_Contains() {
	cache, err := tlru.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{ // insert in user data with user id 1
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	fmt.Println(cache.Contains(1))
	fmt.Println(cache.Contains(2))

	// Output:
	// true
	// false
}

// ExampleLRU_Get shows an example of how Get works and
// how to handle when the key is not found in the cache.
func ExampleLRU_Get() {
	cache, err := tlru.New[int, Member](256)
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	val, ok := cache.Get(1)
	if !ok {
		fmt.Println("[GET] could not find the key in the cache.")
	} else {
		fmt.Printf("[GET] Name : %v | Email : %v\n", val.Name, val.Email)
	}

	val, ok = cache.Get(2)
	if !ok {
		fmt.Println("[GET] could not find the key in the cache.")
	} else {
		fmt.Printf("[GET] Name : %v | Email : %v\n", val.Name, val.Email)
	}
	// Output:
	// [GET] Name : justpranavrs | Email : iliketlru@gmail.com
	// [GET] could not find the key in the cache.
}

// ExampleLRU_GetQuiet shows an example of how GetQuiet works.
// It doesn't disturb the internal state of the cache.
func ExampleLRU_GetQuiet() {
	cache, err := tlru.New[int, Member](2, tlru.WithShards(1))
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRUCore instance: %v", err)
		return
	}

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	cache.Put(3, Member{
		Name:  "welcometotlru",
		Email: "welcometotlru@gmail.com",
	})
	val, ok := cache.Get(1)
	if !ok {
		fmt.Println("[GET] could not find the key in the cache.")
	} else {
		fmt.Printf("[GET] Name : %v | Email : %v\n", val.Name, val.Email)
	}

	val, ok = cache.GetQuiet(3)
	if !ok {
		fmt.Println("[GET] could not find the key in the cache.")
	} else {
		fmt.Printf("[GET] Name : %v | Email : %v\n", val.Name, val.Email)
	}
	cache.Put(2, Member{
		Name:  "tlru",
		Email: "tlruisthebest@gmail.com",
	})
	fmt.Println(cache.Contains(3))

	// Output:
	// [GET] Name : justpranavrs | Email : iliketlru@gmail.com
	// [GET] Name : welcometotlru | Email : welcometotlru@gmail.com
	// false
}

// ExampleLRU_Put shows an example of how Put works and showcases
// the least recently used key getting evicted in a LRU cache.
func ExampleLRU_Put() {
	cache, err := tlru.New[int, Member](256)
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	fmt.Println(cache.Contains(2))
	fmt.Println(cache.Contains(1))

	cache.Put(2, Member{
		Name:  "welcometotlru",
		Email: "welcometotlru@gmail.com",
	})
	fmt.Println(cache.Contains(2))

	cache.Put(3, Member{
		Name:  "justpranavrs",
		Email: "tlruiscool@gmail.com",
	})
	fmt.Println(cache.Contains(4))
	fmt.Println(cache.Contains(3))

	// Output:
	// false
	// true
	// true
	// false
	// true
}

// ExampleLRU_Size shows an example on how Size works. It returns the current size 
// of the LRU cache.
func ExampleLRU_Size() {
	cache, err := tlru.New[int, Member](256)
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	fmt.Println(cache.Size())
	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	fmt.Println(cache.Size())

	// Output:
	// 0
	// 1
}
