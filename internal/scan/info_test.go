package scan

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGetInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	infoInit := GetInfo(ctx)

	if infoInit == (Info{}) {
		t.Error("Test fallido, el struct esta vacío")
	}
	fmt.Println(infoInit)

}
