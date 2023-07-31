package tests

import (
	"adservice/internal/adapters/basic_customer"
	"adservice/internal/adapters/map_repo"
	"context"
	"fmt"
	"strconv"
	"testing"
)

const (
	Users = 50000
	Ads   = 1000000
)

func BenchmarkMapRepo(b *testing.B) {
	ctx := context.Background()
	mapRepo := maprepo.New()
	for i := 0; i < b.N; i++ {
		_, _ = mapRepo.Add(ctx, fmt.Sprint("ad", i), "test ad", 1)
	}
}

func BenchmarkBasicCustomer(b *testing.B) {
	ctx := context.Background()
	basicCustomer := basiccustomer.New()
	for i := 0; i < b.N; i++ {
		_, _ = basicCustomer.CreateByID(ctx, fmt.Sprint("user", i),
			"example"+strconv.Itoa(i)+"@mail.ru", int64(i))
	}
}
