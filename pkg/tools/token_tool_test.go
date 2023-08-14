package tools

import (
	"fmt"
	"testing"
)

func TestGenToken(t *testing.T) {
	token, _ := GenToken(123, 1, 1)
	fmt.Println(token)
}
