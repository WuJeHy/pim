package main

import (
	"fmt"
	"time"
)

type TestData struct {
	groups map[int][]int
}

func main() {

	data := &TestData{
		groups: make(map[int][]int),
	}

	go func() {

		values := []int{1, 2, 2, 3, 4, 5, 8, 7, 4}

		data.groups[123] = append(data.groups[123], values...)

	}()

	go func() {
		time.Sleep(time.Second)
		for i, ints := range data.groups {
			fmt.Println(i, ints)
		}
	}()

	time.Sleep(time.Second * 5)

}
