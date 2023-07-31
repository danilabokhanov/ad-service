package grpc

import (
	"adservice/internal/app"
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (d AdService) CreateAd(ctx context.Context, req *CreateAdRequest) (*AdResponse, error) {
	ad, err := d.a.CreateAd(ctx, req.Title, req.Text, req.UserId)
	if err != nil {
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &AdResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &AdResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &AdResponse{Id: ad.ID,
		Title:        ad.Title,
		Text:         ad.Text,
		AuthorId:     ad.AuthorID,
		Published:    ad.Published,
		CreationDate: timestamppb.New(ad.CreationDate),
		UpdateDate:   timestamppb.New(ad.CreationDate)}, nil
}

func (d AdService) ChangeAdStatus(ctx context.Context, req *ChangeAdStatusRequest) (*AdResponse, error) {
	ad, err := d.a.ChangeAdStatus(ctx, req.AdId, req.UserId, req.Published)
	if err != nil {
		if errors.Is(err, app.ErrNoAccess) {
			return &AdResponse{}, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &AdResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &AdResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &AdResponse{Id: ad.ID,
		Title:        ad.Title,
		Text:         ad.Text,
		AuthorId:     ad.AuthorID,
		Published:    ad.Published,
		CreationDate: timestamppb.New(ad.CreationDate),
		UpdateDate:   timestamppb.New(ad.CreationDate)}, nil
}

func (d AdService) UpdateAd(ctx context.Context, req *UpdateAdRequest) (*AdResponse, error) {
	ad, err := d.a.UpdateAd(ctx, req.AdId, req.UserId, req.Title, req.Text)
	if err != nil {
		if errors.Is(err, app.ErrNoAccess) {
			return &AdResponse{}, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &AdResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &AdResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &AdResponse{Id: ad.ID,
		Title:        ad.Title,
		Text:         ad.Text,
		AuthorId:     ad.AuthorID,
		Published:    ad.Published,
		CreationDate: timestamppb.New(ad.CreationDate),
		UpdateDate:   timestamppb.New(ad.CreationDate)}, nil
}

func (d AdService) DeleteAd(ctx context.Context, req *DeleteAdRequest) (*AdResponse, error) {
	ad, err := d.a.DeleteAd(ctx, req.AdId, req.UserId)
	if err != nil {
		if errors.Is(err, app.ErrNoAccess) {
			return &AdResponse{}, status.Error(codes.PermissionDenied, err.Error())
		}
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &AdResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &AdResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &AdResponse{Id: ad.ID,
		Title:        ad.Title,
		Text:         ad.Text,
		AuthorId:     ad.AuthorID,
		Published:    ad.Published,
		CreationDate: timestamppb.New(ad.CreationDate),
		UpdateDate:   timestamppb.New(ad.CreationDate)}, nil
}

func (d AdService) ListAds(ctx context.Context, req *FilterRequest) (*ListAdResponse, error) {
	f, err := d.a.GetNewFilter(ctx)
	if err != nil {
		return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
	}
	f, err = f.SetAuthor(ctx, req.AuthorId)
	if err != nil {
		return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
	}
	if req.PublishedConfig != PublishedConfig_NotGiven {
		var publishedOnly bool
		if req.PublishedConfig == PublishedConfig_PublishedOnly {
			publishedOnly = true
		}
		f, err = f.SetStatus(ctx, publishedOnly)
		if err != nil {
			return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
		}
	}
	lDate := req.LDate.AsTime().UTC()
	if lDate.Unix() != 0 {
		f, err = f.SetLTime(ctx, lDate)
		if err != nil {
			return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
		}
	}
	rDate := req.RDate.AsTime().UTC()
	if rDate.Unix() != 0 {
		f, err = f.SetRTime(ctx, rDate)
		if err != nil {
			return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
		}
	}
	adp, err := f.GetPattern(ctx)
	if err != nil {
		return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
	}
	ads, err := d.a.GetAllAdsByTemplate(ctx, adp)
	if err != nil {
		return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
	}
	res := ListAdResponse{}
	for _, ad := range ads {
		res.List = append(res.List, &AdResponse{Id: ad.ID,
			Title:        ad.Title,
			Text:         ad.Text,
			AuthorId:     ad.AuthorID,
			Published:    ad.Published,
			CreationDate: timestamppb.New(ad.CreationDate),
			UpdateDate:   timestamppb.New(ad.CreationDate)})
	}
	return &res, nil
}

func (d AdService) GetAdByID(ctx context.Context, req *GetAdRequest) (*AdResponse, error) {
	ad, err := d.a.FindAd(ctx, req.Id)
	if err != nil {
		if errors.Is(err, app.ErrWrongFormat) {
			return &AdResponse{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &AdResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &AdResponse{Id: ad.ID,
		Title:        ad.Title,
		Text:         ad.Text,
		AuthorId:     ad.AuthorID,
		Published:    ad.Published,
		CreationDate: timestamppb.New(ad.CreationDate),
		UpdateDate:   timestamppb.New(ad.CreationDate)}, nil
}

func (d AdService) CreateUser(ctx context.Context, req *UniversalUser) (*UniversalUser, error) {
	u, err := d.a.CreateUserByID(ctx, req.Nickname, req.Email, req.UserId)
	if err != nil {
		if errors.Is(err, app.ErrWrongFormat) {
			return &UniversalUser{}, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, app.ErrNoAccess) {
			return &UniversalUser{}, status.Error(codes.PermissionDenied, err.Error())
		}
		return &UniversalUser{}, status.Error(codes.Internal, err.Error())
	}
	return &UniversalUser{UserId: u.ID, Nickname: u.Nickname, Email: u.Email}, nil
}

func (d AdService) DeleteUserByID(ctx context.Context, req *DeleteUserRequest) (*UniversalUser, error) {
	u, err := d.a.DeleteUserByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &UniversalUser{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &UniversalUser{}, status.Error(codes.Internal, err.Error())
	}
	return &UniversalUser{UserId: u.ID, Nickname: u.Nickname, Email: u.Email}, nil
}

func (d AdService) ChangeUserInfo(ctx context.Context, req *UniversalUser) (*UniversalUser, error) {
	u, err := d.a.ChangeUserInfo(ctx, req.UserId, req.Nickname, req.Email)
	if err != nil {
		if errors.Is(err, app.ErrWrongFormat) || errors.Is(err, app.ErrUserDoesNotExist) {
			return &UniversalUser{}, status.Error(codes.InvalidArgument, err.Error())
		}
		return &UniversalUser{}, status.Error(codes.Internal, err.Error())
	}
	return &UniversalUser{UserId: u.ID, Nickname: u.Nickname, Email: u.Email}, nil
}

func (d AdService) GetAdsByTitle(ctx context.Context, req *AdsByTitleRequest) (*ListAdResponse, error) {
	ads, err := d.a.GetAdsByTitle(ctx, req.Title)
	if err != nil {
		return &ListAdResponse{}, status.Error(codes.Internal, err.Error())
	}
	res := ListAdResponse{}
	for _, ad := range ads {
		res.List = append(res.List, &AdResponse{Id: ad.ID,
			Title:        ad.Title,
			Text:         ad.Text,
			AuthorId:     ad.AuthorID,
			Published:    ad.Published,
			CreationDate: timestamppb.New(ad.CreationDate),
			UpdateDate:   timestamppb.New(ad.CreationDate)})
	}
	return &res, nil
}

func (d AdService) GetUserByID(ctx context.Context, req *GetUserRequest) (*UniversalUser, error) {
	u, isFound, err := d.a.FindUser(ctx, req.Id)
	if err != nil {
		return &UniversalUser{}, status.Error(codes.Internal, err.Error())
	}
	if !isFound {
		return &UniversalUser{}, status.Error(codes.NotFound, "")
	}
	return &UniversalUser{UserId: u.ID, Nickname: u.Nickname, Email: u.Email}, nil
}
