package main

import (
	"fmt"
	"github.com/cornelk/hashmap"
)

func main() {
	h := hashmap.HashMap{}
	var l []int
	h.Set("1234", l)
	if k, ok := h.Get("1234"); ok {
		fmt.Println(k.([]int))
		l = k.([]int)
		l = append(l, 3)
		h.Set("1234", l)
		if k, ok := h.Get("1234"); ok {
			fmt.Println(k.([]int))
		}

	}

	fmt.Println(h)
}
