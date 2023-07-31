package tests

import (
	"adservice/internal/adapters/basic_customer"
	"adservice/internal/adapters/map_repo"
	"context"
	"strconv"
	"testing"
)

const IntBase int = 8

func FuzzMapRepo(f *testing.F) {
	testcases := []int64{0, 1000}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, n int64) {
		mapRepo := maprepo.New()
		ctx := context.Background()
		for i := int64(0); i < n; i += 1 {
			_, _ = mapRepo.Add(ctx, strconv.Itoa(int(i)), "some text", n)
		}
		got, _ := mapRepo.Add(ctx, strconv.Itoa(int(n)), "some text", 1)
		expect := n

		if got != expect {
			t.Errorf("For (%d) Expect: %d, but got: %d", n, expect, got)
		}
	})
}

func genBoolSlice(mask uint8) []bool {
	var s []bool
	for i := 0; i < IntBase; i++ {
		s = append(s, ((mask>>i)&1) == 1)
	}
	return s
}

func FuzzBasicCustomer(f *testing.F) {
	testcases := []uint8{0, 255}

	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, mask uint8) {
		ctx := context.Background()
		s := genBoolSlice(mask)
		basicCustomer := basiccustomer.New()
		for i := 0; i < len(s); i++ {
			if s[i] {
				_, _ = basicCustomer.CreateByID(ctx, "user with id: "+strconv.Itoa(i),
					"example"+strconv.Itoa(i)+"@mair.ru", int64(i))
			}
		}
		for i := 0; i < len(s); i++ {
			_, got := basicCustomer.Find(ctx, int64(i))
			expect := s[i]
			if got != expect {
				t.Errorf("For slice(%v\n) Value: %d, Expect: %t, but got: %t", s, i, expect, got)
				break
			}
		}
	})
}
