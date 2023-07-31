package tests

import (
	"adservice/internal/ads"
	postgrespgx "adservice/internal/ports/pgx"
	"adservice/internal/user"
	"context"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestMySQLConfig struct {
	suite.Suite
	userDB postgrespgx.PGXUsers
	adDB   postgrespgx.PGXRepo
}

func (suite *TestMySQLConfig) SetupTest() {
	suite.adDB = postgrespgx.NewDBRepo()
	err := suite.adDB.DeleteAll(context.Background())
	suite.Assert().NoError(err)

	suite.userDB = postgrespgx.NewDBUsers()
	err = suite.userDB.DeleteAll(context.Background())
	suite.Assert().NoError(err)
}

func (suite *TestMySQLConfig) TearDownTest() {
	_ = suite.userDB.Close()
	_ = suite.adDB.Close()
}

func (suite *TestMySQLConfig) TestBasicUserDB() {
	err := suite.userDB.Insert(context.Background(), "cat", "cat@mail.ru", 3)
	suite.Assert().NoError(err)
	users, err := suite.userDB.SelectById(context.Background(), 3)
	suite.Assert().NoError(err)
	suite.Assert().Equal([]user.User{{ID: 3, Nickname: "cat", Email: "cat@mail.ru"}}, users)
	err = suite.userDB.DeleteById(context.Background(), 3)
	suite.Assert().NoError(err)
	users, err = suite.userDB.SelectAll(context.Background())
	suite.Assert().NoError(err)
	suite.Assert().Equal([]user.User{}, users)
}

func (suite *TestMySQLConfig) TestBasicAdDB() {
	err := suite.userDB.Insert(context.Background(), "cat", "cat@mail.ru", 3)
	suite.Assert().NoError(err)
	ad := ads.Ad{ID: 1, Title: "test title", Text: "test text", AuthorID: 3, Published: false,
		CreationDate: time.UnixMicro(time.Now().UTC().UnixMicro()).UTC(),
		UpdateDate:   time.UnixMicro(time.Now().UTC().UnixMicro()).UTC()}
	err = suite.adDB.Insert(context.Background(), ad)
	suite.Assert().NoError(err)
	adverts, err := suite.adDB.SelectById(context.Background(), 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal([]ads.Ad{ad}, adverts)
}

func TestMySQL(t *testing.T) {
	suite.Run(t, new(TestMySQLConfig))
}
