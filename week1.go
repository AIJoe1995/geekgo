package main

import (
	"errors"
	"fmt"
)

func ShrinkSlice[T any](slice []T) []T {
	prevcap := cap(slice)
	newcap := cap(slice)
	if len(slice) < cap(slice)/2 {
		newcap = cap(slice) / 2
	}
	if prevcap == newcap {
		return slice
	} else {
		fmt.Printf("shrink capacity")
		res := make([]T, 0, newcap)
		res = append(res, slice...)
		return res
	}

}

func deleteAt[T any](slice []T, idx int) ([]T, error) {

	if idx < 0 || idx >= len(slice) {
		return nil, errors.New("index error")
	}
	head_slice := slice[:idx]
	tail_slice := slice[idx+1:]
	res := make([]T, 0, cap(slice))
	res = append(res, head_slice...)
	res = append(res, tail_slice...)

	res = ShrinkSlice(res)
	return res, nil

}

// 实现泛型方法

// 支持缩容 设计缩容机制

func myprint[T any](res []T, err error) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v, len: %d, cap: %d \n", res, len(res), cap(res))
}

func main() {
	res := make([]int, 8, 8)
	for i := 0; i < len(res); i++ {
		res[i] = i
	}
	// err := nil cannot assign nil without explicit type
	myprint[int](res, nil)

	res, err := deleteAt(res, 0)
	myprint[int](res, err)

	res, err = deleteAt(res, 0)
	myprint[int](res, err)

	res, err = deleteAt(res, 0)
	myprint[int](res, err)

	res, err = deleteAt(res, 0)
	myprint[int](res, err)

	res, err = deleteAt(res, 0)
	myprint[int](res, err)

}
