package tool

import (
	"fmt"
	"testing"
	"time"
)

func TestGetRandomNM(t *testing.T) {
	for i:=0;i<30;i++{
		time.Sleep(time.Millisecond*10)
		fmt.Printf("%d ", GetRandomNM(1, 5))
	}
}
