package main

import (
	"errors"
	"testing"
)

func TestCheck(t *testing.T) {
	e := errors.New("dummy error")
	check(nil)
	assertCheckPanic(t, check, e)
}

func assertCheckPanic(t *testing.T, f func(error), e error) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f(e)
}
