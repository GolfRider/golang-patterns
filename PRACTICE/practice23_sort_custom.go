package main

import (
	"cmp"
	"fmt"
	"slices"
)

type User struct {
	Name string
	Age  int
}

func problem23() {
	users := []User{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 25},
		{"Dan", 35},
	}

	// Custom Sort Logic
	slices.SortFunc(users, func(a, b User) int {
		// Compare Age first
		if n := cmp.Compare(a.Age, b.Age); n != 0 {
			return n
		}
		// If ages are equal, compare Name
		return cmp.Compare(a.Name, b.Name)
	})

	fmt.Println(users)
}
