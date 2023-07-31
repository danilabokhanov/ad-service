package tests

import (
	basicfilter "adservice/internal/adapters/basic_filter"
	"adservice/internal/adapters/db_customer"
	"adservice/internal/adapters/db_repo"
	"adservice/internal/ports/pgx"
	"context"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
	"testing"
	"time"

	"adservice/internal/app"
	grpcPort "adservice/internal/ports/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	ErrorBadRequest       = status.Error(codes.InvalidArgument, app.ErrWrongFormat.Error())
	ErrorUserDoesNotExist = status.Error(codes.InvalidArgument, app.ErrUserDoesNotExist.Error())
	ErrorForbidden        = status.Error(codes.PermissionDenied, app.ErrNoAccess.Error())
	ErrorNotFound         = status.Error(codes.NotFound, "")
	ErrorInternal         = status.Error(codes.Internal, app.ErrApp.Error())
)

type TestConfig struct {
	suite.Suite
	client grpcPort.AdServiceClient
	ctx    context.Context
	lis    *bufconn.Listener
	srv    *grpc.Server
	cancel context.CancelFunc
	conn   *grpc.ClientConn

	adDB   postgrespgx.PGXRepo
	userDB postgrespgx.PGXUsers
}

func (suite *TestConfig) SetupSuite() {
	suite.adDB = postgrespgx.NewDBRepo()
	err := suite.adDB.DeleteAll(context.Background())
	suite.Assert().NoError(err)

	suite.userDB = postgrespgx.NewDBUsers()
	err = suite.userDB.DeleteAll(context.Background())
	suite.Assert().NoError(err)
}

func (suite *TestConfig) SetupTest() {
	suite.lis = bufconn.Listen(1024 * 1024)

	suite.srv = grpc.NewServer(grpc.ChainUnaryInterceptor(grpcPort.UnaryInterceptor, grpcPort.RecoveryInterceptor))

	//svc := grpcPort.NewService(app.NewApp(maprepo.New(), basiccustomer.New(), basicfilter.New()))
	svc := grpcPort.NewService(app.NewApp(dbrepo.New(&suite.adDB), dbcustomer.New(&suite.userDB), basicfilter.New()))
	grpcPort.RegisterAdServiceServer(suite.srv, svc)

	go func() {
		srv := suite.srv
		lis := suite.lis
		_ = srv.Serve(lis)
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return suite.lis.Dial()
	}

	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Second)

	var err error
	suite.conn, err = grpc.DialContext(suite.ctx, "", grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	suite.Assert().NoError(err, "grpc.DialContext")

	suite.client = grpcPort.NewAdServiceClient(suite.conn)
}

func (suite *TestConfig) TearDownTest() {
	_ = suite.lis.Close()
	suite.srv.Stop()
	suite.cancel()
	_ = suite.conn.Close()

	_ = suite.adDB.DeleteAll(context.Background())
	_ = suite.userDB.DeleteAll(context.Background())
}

func (suite *TestConfig) TearDownSuite() {
	_ = suite.adDB.Close()
	_ = suite.userDB.Close()
}

func (suite *TestConfig) TestGRPCCreateUser() {
	res, err := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	suite.Assert().NoError(err, "suite.client.CreateUser")
	suite.Assert().Equal("Tom", res.Nickname)
	suite.Assert().Equal("example@mail.com", res.Email)
	suite.Assert().Equal(int64(3), res.UserId)

	_, err = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "abc", Email: "cat@mail.com", UserId: 3})
	suite.Assert().ErrorIs(err, ErrorForbidden)

	res, err = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "qwerty", Email: "qwerty@mail.com", UserId: 5})
	suite.Assert().NoError(err, "suite.client.CreateUser")
	suite.Assert().Equal("qwerty", res.Nickname)
	suite.Assert().Equal("qwerty@mail.com", res.Email)
	suite.Assert().Equal(int64(5), res.UserId)
}

func (suite *TestConfig) TestGRPCCreateAd() {
	a, err := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	suite.Assert().NoError(err, "suite.client.CreateUser")

	res, err := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: a.UserId})
	suite.Assert().NoError(err, "suite.client.CreateAd")
	suite.Assert().Equal("cat", res.Title)
	suite.Assert().Equal("text", res.Text)
	suite.Assert().Equal(a.UserId, res.AuthorId)
	suite.Assert().Equal(false, res.Published)

	_, err = suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: 5})
	suite.Assert().ErrorIs(err, ErrorUserDoesNotExist)
}

func (suite *TestConfig) TestGRPCChangeAdStatus() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err := suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id, UserId: b.UserId, Published: ad.Published})
	suite.Assert().ErrorIs(err, ErrorForbidden)
	_, err = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id + 1, UserId: ad.AuthorId, Published: ad.Published})
	suite.Assert().ErrorIs(err, ErrorBadRequest)

	updatedAd, err := suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id, UserId: ad.AuthorId, Published: true})
	suite.Assert().NoError(err, "suite.client.ChangeAdStatus")
	suite.Assert().Equal(ad.Title, updatedAd.Title)
	suite.Assert().Equal(ad.Text, updatedAd.Text)
	suite.Assert().Equal(ad.AuthorId, updatedAd.AuthorId)
	suite.Assert().Equal(true, updatedAd.Published)
}

func (suite *TestConfig) TestGRPCUpdateAd() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err := suite.client.UpdateAd(suite.ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id, Title: "new title", Text: "new text", UserId: b.UserId})
	suite.Assert().ErrorIs(err, ErrorForbidden)
	_, err = suite.client.UpdateAd(suite.ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id + 1, Title: "new title", Text: "new text", UserId: ad.AuthorId})
	suite.Assert().ErrorIs(err, ErrorBadRequest)

	updatedAd, err := suite.client.UpdateAd(suite.ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id, Title: "new title", Text: "new text", UserId: ad.AuthorId})
	suite.Assert().NoError(err, "suite.client.UpdateAd")
	suite.Assert().Equal("new title", updatedAd.Title)
	suite.Assert().Equal("new text", updatedAd.Text)
	suite.Assert().Equal(ad.AuthorId, updatedAd.AuthorId)
	suite.Assert().Equal(ad.Published, updatedAd.Published)
}

func (suite *TestConfig) TestGRPCDeleteAd() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err := suite.client.DeleteAd(suite.ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id, UserId: b.UserId})
	suite.Assert().ErrorIs(err, ErrorForbidden)
	_, err = suite.client.DeleteAd(suite.ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id + 1, UserId: ad.AuthorId})
	suite.Assert().ErrorIs(err, ErrorBadRequest)

	resp, err := suite.client.DeleteAd(suite.ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id, UserId: ad.AuthorId})
	suite.Assert().NoError(err, "suite.client.DeleteAd")
	suite.Assert().Equal(ad.Title, resp.Title)
	suite.Assert().Equal(ad.Text, resp.Text)
	suite.Assert().Equal(ad.AuthorId, resp.AuthorId)
	suite.Assert().Equal(ad.Published, resp.Published)
}

func (suite *TestConfig) TestGRPCGetAdByID() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	ad, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err := suite.client.GetAdByID(suite.ctx, &grpcPort.GetAdRequest{
		Id: ad.Id + 1})
	suite.Assert().ErrorIs(err, ErrorBadRequest)

	resp, err := suite.client.GetAdByID(suite.ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	suite.Assert().NoError(err, "suite.client.GetAdByID")
	suite.Assert().Equal(ad.Title, resp.Title)
	suite.Assert().Equal(ad.Text, resp.Text)
	suite.Assert().Equal(ad.AuthorId, resp.AuthorId)
	suite.Assert().Equal(ad.Published, resp.Published)
}

func (suite *TestConfig) TestGRPCDeleteUserByID() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	ad, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err := suite.client.DeleteUserByID(suite.ctx, &grpcPort.DeleteUserRequest{
		Id: a.UserId + 1})
	suite.Assert().ErrorIs(err, ErrorUserDoesNotExist)
	_, err = suite.client.GetAdByID(suite.ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	suite.Assert().NoError(err, "suite.client.GetAdByID")

	resp, err := suite.client.DeleteUserByID(suite.ctx, &grpcPort.DeleteUserRequest{
		Id: a.UserId})
	suite.Assert().NoError(err, "suite.client.DeleteAd")
	suite.Assert().Equal(a.Nickname, resp.Nickname)
	suite.Assert().Equal(a.Email, resp.Email)
	suite.Assert().Equal(a.UserId, resp.UserId)

	_, err = suite.client.GetAdByID(suite.ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	suite.Assert().ErrorIs(err, ErrorBadRequest)
}

func (suite *TestConfig) TestGRPCChangeUserInfo() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	_, err := suite.client.ChangeUserInfo(suite.ctx, &grpcPort.UniversalUser{
		UserId:   a.UserId + 1,
		Nickname: "Qwerty",
		Email:    "qwerty@mail.ru",
	})
	suite.Assert().ErrorIs(err, ErrorUserDoesNotExist)

	resp, err := suite.client.ChangeUserInfo(suite.ctx, &grpcPort.UniversalUser{
		UserId:   a.UserId,
		Nickname: "Qwerty",
		Email:    "qwerty@mail.ru",
	})
	suite.Assert().NoError(err, "suite.client.ChangeUserInfo")
	suite.Assert().Equal("Qwerty", resp.Nickname)
	suite.Assert().Equal("qwerty@mail.ru", resp.Email)
	suite.Assert().Equal(a.UserId, resp.UserId)
}

func (suite *TestConfig) TestGRPCGetAdsByTitle() {
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	a, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: 3})
	time.Sleep(time.Millisecond)
	b, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "abacaba", Text: "12345", UserId: 3})
	_, _ = suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: 3})

	resp, err := suite.client.GetAdsByTitle(suite.ctx, &grpcPort.AdsByTitleRequest{Title: "aba"})
	suite.Assert().NoError(err, "suite.client.GetAdsByTitle")
	suite.Assert().Equal(a.Title, resp.List[0].Title)
	suite.Assert().Equal(a.Text, resp.List[0].Text)
	suite.Assert().Equal(a.AuthorId, resp.List[0].AuthorId)
	suite.Assert().Equal(a.Published, resp.List[0].Published)
	suite.Assert().Equal(b.Title, resp.List[1].Title)
	suite.Assert().Equal(b.Text, resp.List[1].Text)
	suite.Assert().Equal(b.AuthorId, resp.List[1].AuthorId)
	suite.Assert().Equal(b.Published, resp.List[1].Published)
}

func (suite *TestConfig) TestGRPCGetUserByID() {
	a, _ := suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	_, err := suite.client.GetUserByID(suite.ctx, &grpcPort.GetUserRequest{
		Id: a.UserId + 1,
	})
	suite.Assert().ErrorIs(err, ErrorNotFound)

	resp, err := suite.client.GetUserByID(suite.ctx, &grpcPort.GetUserRequest{
		Id: a.UserId,
	})
	suite.Assert().NoError(err, "suite.client.GetUserByID")
	suite.Assert().Equal(a.Nickname, resp.Nickname)
	suite.Assert().Equal(a.Email, resp.Email)
	suite.Assert().Equal(a.UserId, resp.UserId)
}

func (suite *TestConfig) TestGRPCDefaultFilter() {
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 123,
		Nickname: "nickname", Email: "example@mail.com"})

	resp, err := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 123,
		Title: "hello", Text: "world"})
	suite.Assert().NoError(err, "suite.client.CreateAd")

	publishedAd, err := suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 123, AdId: resp.Id, Published: true})
	suite.Assert().NoError(err, "suite.client.ChangeAdStatus")

	_, err = suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 123,
		Title: "best cat", Text: "not for sale"})
	suite.Assert().NoError(err, "suite.client.CreateAd")

	ads, err := suite.client.ListAds(suite.ctx, &grpcPort.FilterRequest{})
	suite.Assert().NoError(err, "suite.client.ListAds")
	suite.Assert().Len(ads.List, 1)
	suite.Assert().Equal(ads.List[0].Id, publishedAd.Id)
	suite.Assert().Equal(ads.List[0].Title, publishedAd.Title)
	suite.Assert().Equal(ads.List[0].Text, publishedAd.Text)
	suite.Assert().Equal(ads.List[0].AuthorId, publishedAd.AuthorId)
	suite.Assert().True(ads.List[0].Published)
}

func (suite *TestConfig) TestGRPCFilterByAuthor() {
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 3,
		Nickname: "nickname", Email: "example@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 5,
		Nickname: "cat", Email: "cat@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 7,
		Nickname: "aba", Email: "caba@mail.com"})

	a, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "aba", Text: "caba"})
	time.Sleep(time.Millisecond)
	b, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "bab", Text: "abac"})
	c, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 5,
		Title: "foo", Text: "bar"})
	d, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 7,
		Title: "alpha", Text: "beta"})

	a, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: a.Id, Published: true})
	b, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: b.Id, Published: true})
	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 5, AdId: c.Id, Published: true})
	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 7, AdId: d.Id, Published: true})

	ads, err := suite.client.ListAds(suite.ctx, &grpcPort.FilterRequest{AuthorId: 3})
	suite.Assert().NoError(err)
	suite.Assert().Equal(ads.List[0].Id, a.Id)
	suite.Assert().Equal(ads.List[0].Title, a.Title)
	suite.Assert().Equal(ads.List[0].Text, a.Text)
	suite.Assert().Equal(ads.List[0].AuthorId, a.AuthorId)
	suite.Assert().Equal(ads.List[0].Published, a.Published)

	suite.Assert().Equal(ads.List[1].Id, b.Id)
	suite.Assert().Equal(ads.List[1].Title, b.Title)
	suite.Assert().Equal(ads.List[1].Text, b.Text)
	suite.Assert().Equal(ads.List[1].AuthorId, b.AuthorId)
	suite.Assert().Equal(ads.List[1].Published, b.Published)
}

func (suite *TestConfig) TestGRPCFilterByPublishedOnly() {
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 3,
		Nickname: "nickname", Email: "example@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 5,
		Nickname: "cat", Email: "cat@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 7,
		Nickname: "aba", Email: "caba@mail.com"})

	a, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "aba", Text: "caba"})
	b, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "bab", Text: "abac"})
	time.Sleep(time.Millisecond)
	c, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 5,
		Title: "foo", Text: "bar"})
	d, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 7,
		Title: "alpha", Text: "beta"})

	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: a.Id, Published: false})
	b, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: b.Id, Published: true})
	c, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 5, AdId: c.Id, Published: true})
	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 7, AdId: d.Id, Published: false})

	ads, err := suite.client.ListAds(suite.ctx,
		&grpcPort.FilterRequest{PublishedConfig: grpcPort.PublishedConfig_PublishedOnly})
	suite.Assert().NoError(err)
	suite.Assert().Equal(ads.List[0].Id, b.Id)
	suite.Assert().Equal(ads.List[0].Title, b.Title)
	suite.Assert().Equal(ads.List[0].Text, b.Text)
	suite.Assert().Equal(ads.List[0].AuthorId, b.AuthorId)
	suite.Assert().Equal(ads.List[0].Published, b.Published)

	suite.Assert().Equal(ads.List[1].Id, c.Id)
	suite.Assert().Equal(ads.List[1].Title, c.Title)
	suite.Assert().Equal(ads.List[1].Text, c.Text)
	suite.Assert().Equal(ads.List[1].AuthorId, c.AuthorId)
	suite.Assert().Equal(ads.List[1].Published, c.Published)
}

func (suite *TestConfig) TestGRPCFilterByTime() {
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 3,
		Nickname: "nickname", Email: "example@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 5,
		Nickname: "cat", Email: "cat@mail.com"})
	_, _ = suite.client.CreateUser(suite.ctx, &grpcPort.UniversalUser{UserId: 7,
		Nickname: "aba", Email: "caba@mail.com"})

	a, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "aba", Text: "caba"})
	time.Sleep(time.Millisecond)
	lTm := timestamppb.New(time.Now())
	time.Sleep(time.Millisecond)
	b, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "bab", Text: "abac"})
	time.Sleep(time.Millisecond)
	c, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 5,
		Title: "foo", Text: "bar"})
	time.Sleep(time.Millisecond)
	rTm := timestamppb.New(time.Now())
	time.Sleep(time.Millisecond)
	d, _ := suite.client.CreateAd(suite.ctx, &grpcPort.CreateAdRequest{UserId: 7,
		Title: "alpha", Text: "beta"})

	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: a.Id, Published: true})
	b, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: b.Id, Published: true})
	c, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 5, AdId: c.Id, Published: true})
	_, _ = suite.client.ChangeAdStatus(suite.ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 7, AdId: d.Id, Published: true})

	ads, err := suite.client.ListAds(suite.ctx, &grpcPort.FilterRequest{LDate: lTm, RDate: rTm})
	suite.Assert().NoError(err)
	suite.Assert().Equal(ads.List[0].Id, b.Id)
	suite.Assert().Equal(ads.List[0].Title, b.Title)
	suite.Assert().Equal(ads.List[0].Text, b.Text)
	suite.Assert().Equal(ads.List[0].AuthorId, b.AuthorId)
	suite.Assert().Equal(ads.List[0].Published, b.Published)

	suite.Assert().Equal(ads.List[1].Id, c.Id)
	suite.Assert().Equal(ads.List[1].Title, c.Title)
	suite.Assert().Equal(ads.List[1].Text, c.Text)
	suite.Assert().Equal(ads.List[1].AuthorId, c.AuthorId)
	suite.Assert().Equal(ads.List[1].Published, c.Published)
}

func TestTestConfig(t *testing.T) {
	suite.Run(t, new(TestConfig))
}
