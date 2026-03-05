package main

import "fmt"

func problem21() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	panic("this is forced panic")
}
