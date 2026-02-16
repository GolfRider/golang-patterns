package main

import (
	"errors"
	"fmt"
	"time"
)

// 1 - Error Handling
func checkError() {
	err := doWork()
	var ae *AppErrror
	if err != nil && errors.As(err, &ae) {
		fmt.Println("Error=>>", ae.Msg, err.Error())
	}
}

func doWork() error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("oops some error: %w", &AppErrror{ID: 101, Msg: "This is an error"})
}

type AppErrror struct {
	ID  int
	Msg string
}

func (ap *AppErrror) Error() string {
	return ap.Msg
}
