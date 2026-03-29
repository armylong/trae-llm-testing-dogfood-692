package demo

import (
	"context"
	"fmt"
	"testing"
)

func TestDemoBusiness_SetMessage(t *testing.T) {
	res, err := DemoBusiness.SetMessage(context.Background(), "longlonglong2")
	fmt.Println(res, err)
}

func TestDemoBusiness_GetMessage(t *testing.T) {
	res, err := DemoBusiness.GetMessage(context.Background())
	fmt.Println(res, err)
}
