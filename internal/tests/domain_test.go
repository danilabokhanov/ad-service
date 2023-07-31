package tests

import (
	"adservice/internal/adapters/basic_customer"
	"adservice/internal/adapters/basic_filter"
	"adservice/internal/adapters/map_repo"
	"adservice/internal/app"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeStatusAdOfAnotherUser(t *testing.T) {
	client := getTestClient(app.NewApp(maprepo.New(), basiccustomer.New(), basicfilter.New()))

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	resp, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)

	_, _ = client.createUser(100, "qwerty", "abcde@mail.com")
	_, err = client.changeAdStatus(100, resp.Data.ID, true)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpdateAdOfAnotherUser(t *testing.T) {
	client := getTestClient(app.NewApp(maprepo.New(), basiccustomer.New(), basicfilter.New()))

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	resp, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)

	_, _ = client.createUser(100, "qwerty", "abcde@mail.com")

	_, err = client.updateAd(100, resp.Data.ID, "title", "text")
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestCreateAd_ID(t *testing.T) {
	client := getTestClient(app.NewApp(maprepo.New(), basiccustomer.New(), basicfilter.New()))

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	resp, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)
	assert.Equal(t, resp.Data.ID, int64(0))

	resp, err = client.createAd(123, "hello", "world")
	assert.NoError(t, err)
	assert.Equal(t, resp.Data.ID, int64(1))

	resp, err = client.createAd(123, "hello", "world")
	assert.NoError(t, err)
	assert.Equal(t, resp.Data.ID, int64(2))
}
