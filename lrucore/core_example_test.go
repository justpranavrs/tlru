// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore_test

import (
	"fmt"
	"strconv"

	"github.com/justpranavrs/tlru/lrucore"
)

// Member is the type of the value stored in the cache.
type Member struct {
	Name  string
	Email string
}

// ExampleCore shows a small example of how to initialize a Core instance and
// do basic operations like Put, Size, Peek and Capacity.
func ExampleCore() {
	cache, err := lrucore.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize Core instance: %v", err)
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

// ExampleCore_GetMany shows an example on how GetMany works.
func ExampleCore_GetMany() {
	cache, err := lrucore.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	keys := make([]int, 40)
	for i := 0; i < 36; i++ {
		keys[i] = i
		name := "user_" + strconv.Itoa(i)
		cache.Put(i, Member{
			Name:  name,
			Email: name + "@gmail.com",
		})
	}
	for i := 36; i < 40; i++ {
		keys[i] = i
	}

	values := make([]Member, 40)
	exists := make([]bool, 40)

	if err := cache.GetMany(keys, values, exists); err != nil {
		fmt.Printf("[GET-MANY] error: %v", err)
	}
	if exists[12] {
		val := values[12]
		fmt.Printf("[GET-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Printf("[GET-MANY] Key %d is not present in the cache", keys[12])
	}

	if exists[34] {
		val := values[34]
		fmt.Printf("[GET-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Printf("[GET-MANY] Key %d is not present in the cache", keys[34])
	}

	if exists[38] {
		val := values[38]
		fmt.Printf("[GET-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Printf("[GET-MANY] Key %d is not present in the cache", keys[38])
	}

	// Output:
	// [GET-MANY] Name : user_12 | Email : user_12@gmail.com
	// [GET-MANY] Name : user_34 | Email : user_34@gmail.com
	// [GET-MANY] Key 38 is not present in the cache
}

// ExampleCore_PutMany shows an example on how PutMany works.
func ExampleCore_PutMany() {
	cache, err := lrucore.New[int, Member](256) // create a lru instance
	if err != nil {
		fmt.Printf("[ERROR] could not initialize LRU instance: %v", err)
		return
	}

	keys := make([]int, 40)
	values := make([]Member, 40)
	for i := 0; i < 36; i++ {
		keys[i] = i
		name := "user_" + strconv.Itoa(i)

		values[i] = Member{
			Name:  name,
			Email: name + "@gmail.com",
		}
	}
	if err := cache.PutMany(keys, values); err != nil {
		fmt.Printf("[PUT-MANY] error: %v", err)
	}

	val, ok := cache.Peek(12)
	if ok {
		fmt.Printf("[PUT-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Printf("[PUT-MANY] Key %d is not present in the cache", keys[12])
	}

	val, ok = cache.Peek(23)
	if ok {
		fmt.Printf("[PUT-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Printf("[PUT-MANY] Key %d is not present in the cache", keys[23])
	}

	val, ok = cache.Peek(37)
	if ok {
		fmt.Printf("[PUT-MANY] Name : %v | Email : %v\n", val.Name, val.Email)
	} else {
		fmt.Println("[PUT-MANY] Key 37 is not present in the cache")
	}

	// Output:
	// [PUT-MANY] Name : user_12 | Email : user_12@gmail.com
	// [PUT-MANY] Name : user_23 | Email : user_23@gmail.com
	// [PUT-MANY] Key 37 is not present in the cache
}
