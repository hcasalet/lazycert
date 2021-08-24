package main

import (
	hm "github.com/cornelk/hashmap"
	"github.com/hcasalet/lazycert/dump/lc"
	"log"
)

func foo() {
	h := &hm.HashMap{}
	h.Insert("abcd", lc.BlockInfo{
		LogID: 15,
	})
	val, ok := h.Get("abcd")
	log.Println(val, ok)
	v := val.(lc.BlockInfo)
	v.LogID = 20
	log.Println(val, ok)

	h.Insert(123, "abcd")
	val, ok = h.Get(123)
	log.Println(val, ok)
}

func main() {
	foo()
}
