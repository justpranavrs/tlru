// // Copyright (c) 2026 Pranav R S All rights reserved.
// // Use of this source code is governed by a BSD-style
// // license that can be found in the LICENSE file.

package tlru_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/justpranavrs/tlru"
	"github.com/justpranavrs/tlru/core"
)

// Member is the type of the value stored in the cache.
type Member struct {
	Name  string
	Email string
}

// ExampleLRU shows a small example of how to initialize a LRU instance and
// do basic operations like Put, Size, Peek and Capacity.
func ExamplePoolLRU() {
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

	_, ok := cache.Peek(2) // reports whether key 2 is present in the cache
	fmt.Println(ok)

	_, ok = cache.Peek(1)
	fmt.Println(ok) // reports whether key 1 is present in the cache

	fmt.Println(cache.Capacity()) // reports the maximum capacity of the cache.

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

// ExamplePoolLRU_Capacity shows an example of how Capacity works.
func ExamplePoolLRU_Capacity() {
	cache, err := tlru.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	fmt.Println(cache.Capacity())

	// Output:
	// 256
}

// ExamplePoolLRU_Delete shows an example of how Delete works.
func ExamplePoolLRU_Delete() {
	cache, err := tlru.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	cache.Put(2, Member{
		Name:  "welcometotlru",
		Email: "welcometotlru@gmail.com",
	})

	val, ok := cache.Delete(1) // delete key 1 from the cache
	if !ok {
		fmt.Println("[DELETE] could not find the key in the cache.")
	} else {
		fmt.Printf("[DELETE] Name : %v | Email : %v\n", val.Name, val.Email)
	}
	val, ok = cache.Delete(3) // delete key 3 from the cache
	if !ok {                  // key is not present in the cache
		fmt.Println("[DELETE] could not find the key in the cache.")
	} else {
		fmt.Printf("[DELETE] Name : %v | Email : %v\n", val.Name, val.Email)
	}

	val, ok = cache.Peek(1) // check whether key 1 is in the cache.
	fmt.Println(ok)

	// Output:
	// [DELETE] Name : justpranavrs | Email : iliketlru@gmail.com
	// [DELETE] could not find the key in the cache.
	// false
}

// ExamplePoolLRU_Flush shows an example of how Flush works.
func ExamplePoolLRU_Flush() {
	cache, err := tlru.New[int, Member](2560, tlru.WithShards(64)) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	for i := 0; i < 36; i++ {
		name := "user_" + strconv.Itoa(i)
		cache.Put(i, Member{
			Name:  name,
			Email: name + "@gmail.com",
		})
	}
	fmt.Println(cache.Size())

	cache.Flush() // empties the cache.
	fmt.Println(cache.Size())

	// Output:
	// 36
	// 0
}

// ExamplePoolLRU_Get shows an example of how Get works and
// how to handle when the key is not found in the cache.
func ExamplePoolLRU_Get() {
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

// ExamplePoolLRU_Peek shows an example of how Peek works.
// It doesn't disturb the internal state of the cache.
func ExamplePoolLRU_Peek() {
	cache, err := tlru.New[int, Member](2, tlru.WithShards(1))
	if err != nil {
		fmt.Printf("[ERROR] could not initialize Core instance: %v", err)
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

	val, ok = cache.Peek(3)
	if !ok {
		fmt.Println("[GET] could not find the key in the cache.")
	} else {
		fmt.Printf("[GET] Name : %v | Email : %v\n", val.Name, val.Email)
	}
	cache.Put(2, Member{
		Name:  "tlru",
		Email: "tlruisthebest@gmail.com",
	})
	_, ok = cache.Peek(3)
	fmt.Println(ok)

	// Output:
	// [GET] Name : justpranavrs | Email : iliketlru@gmail.com
	// [GET] Name : welcometotlru | Email : welcometotlru@gmail.com
	// false
}

// ExamplePoolLRU_Put shows an example of how Put works.
func ExamplePoolLRU_Put() {
	cache, err := tlru.New[int, Member](256)
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	_, ok := cache.Peek(2)
	fmt.Println(ok)

	_, ok = cache.Peek(1)
	fmt.Println(ok)

	cache.Put(2, Member{
		Name:  "welcometotlru",
		Email: "welcometotlru@gmail.com",
	})

	_, ok = cache.Peek(2)
	fmt.Println(ok)

	cache.Put(3, Member{
		Name:  "justpranavrs",
		Email: "tlruiscool@gmail.com",
	})

	_, ok = cache.Peek(4)
	fmt.Println(ok)

	_, ok = cache.Peek(3)
	fmt.Println(ok)

	// Output:
	// false
	// true
	// true
	// false
	// true
}

// ExamplePoolLRU_Shards shows an example of how Shards works.
func ExamplePoolLRU_Shards() {
	cache, err := tlru.New[int, Member](256, tlru.WithShards(64)) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}
	fmt.Println(cache.Shards())

	// Output:
	// 64
}

// ExamplePoolLRU_Size shows an example on how Size works. It returns the current size
// of the LRU cache.
func ExamplePoolLRU_Size() {
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

// ExamplePoolLRU_Upsert shows an example of how Upsert works.
func ExamplePoolLRU_Upsert() {
	cache, err := tlru.New[int, Member](2, tlru.WithShards(1))
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	state, _ := cache.Upsert(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	if state == core.UpsertAddNoEviction {
		fmt.Println("[UPSERT] : Add without Eviction")
	}

	_, ok := cache.Peek(2)
	fmt.Println(ok)

	_, ok = cache.Peek(1)
	fmt.Println(ok)

	state, _ = cache.Upsert(2, Member{
		Name:  "welcometotlru",
		Email: "welcometotlru@gmail.com",
	})
	if state == core.UpsertAddNoEviction {
		fmt.Println("[UPSERT] : Add without Eviction")
	}

	_, ok = cache.Peek(2)
	fmt.Println(ok)

	state, val := cache.Upsert(3, Member{
		Name:  "justpranavrs",
		Email: "tlruiscool@gmail.com",
	})
	if state == core.UpsertAddWithEviction {
		fmt.Println("[UPSERT] : Add on Eviction")
		fmt.Printf("[UPSERT] Name : %v | Email : %v\n", val.Name, val.Email)
	}

	_, ok = cache.Peek(1)
	fmt.Println(ok)

	_, ok = cache.Peek(3)
	fmt.Println(ok)

	state, val = cache.Upsert(3, Member{
		Name:  "justpranavrs",
		Email: "jprs-tlru@gmail.com",
	})
	if state == core.UpsertReplace {
		fmt.Println("[UPSERT] : Value Replaced")
		fmt.Printf("[UPSERT] Name : %v | Email : %v\n", val.Name, val.Email)
	}

	// Output:
	// [UPSERT] : Add without Eviction
	// false
	// true
	// [UPSERT] : Add without Eviction
	// true
	// [UPSERT] : Add on Eviction
	// [UPSERT] Name : justpranavrs | Email : iliketlru@gmail.com
	// false
	// true
	// [UPSERT] : Value Replaced
	// [UPSERT] Name : justpranavrs | Email : tlruiscool@gmail.com
}

// ExamplePoolTLRU_Close shows an example of how Close works.
func ExamplePoolTLRU_Close() {
	cache, err := tlru.NewWithTTL[int, Member](256, 24*time.Hour) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}
	defer cache.Close() // Close safely shuts down the internal clock.

	cache.Put(1, Member{
		Name:  "justpranavrs",
		Email: "iliketlru@gmail.com",
	})
	fmt.Println(cache.Size())

	// Output:
	// 1
}
