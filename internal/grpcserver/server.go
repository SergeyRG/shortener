package grpcserver

import (
	"context"
	"errors"

	pb "github.com/SergeyRG/shortener/api/proto/shortener/v1"
	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShortenerService struct {
	pb.UnimplementedShortenerServiceServer
	svc service.URLServiceInterface
}

func NewShortenerService(svc service.URLServiceInterface) *ShortenerService {
	return &ShortenerService{svc: svc}
}

func (s *ShortenerService) ShortenURL(
	ctx context.Context,
	in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	url := in.GetUrl()
	id, err := s.svc.AddShortURL(ctx, url, userID)
	if err != nil && !errors.Is(err, service.ErrConflict) {
		logging.Logger.Error("cant add short URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "cant add short URL")
	}

	isErrConflict := errors.Is(err, service.ErrConflict)

	shortURL, err := s.svc.MakeShortURLByID(ctx, id)
	if err != nil {
		logging.Logger.Error("cant make short URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "cant make short URL")
	}

	if isErrConflict {
		respStatus := status.New(codes.AlreadyExists, "Already exists")

		details := &errdetails.ErrorInfo{
			Reason:   "URL already exists",
			Domain:   "ShortenerService/ShortenURL",
			Metadata: map[string]string{"short_url": shortURL},
		}

		stWithDetails, err := respStatus.WithDetails(details)

		if err != nil {
			logging.Logger.Error("cant pack short URL in error details", zap.Error(err))
			return nil, status.Error(codes.Internal, "cant pack short URL in error details")
		}

		return nil, stWithDetails.Err()
	}

	responce := pb.URLShortenResponse_builder{
		Result: shortURL,
	}.Build()

	return responce, nil
}

func (s *ShortenerService) ExpandURL(
	ctx context.Context,
	in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {

	id := in.GetId()
	url, err := s.svc.GetOriginalURLByID(ctx, id)

	if err != nil {
		logging.Logger.Error("cant get original URL by ID", zap.Error(err))
		return nil, status.Error(codes.NotFound, "cant get original URL by ID")
	}

	if url.DeletedFlag {
		logging.Logger.Error("original URL is deleted")
		return nil, status.Error(codes.NotFound, "original URL is deleted")
	}

	return pb.URLExpandResponse_builder{
		Result: url.OriginURL,
	}.Build(), nil
}

func (s *ShortenerService) ListUserURLs(
	ctx context.Context,
	in *pb.ListUserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	userURL, err := s.svc.GetURLByUserID(ctx, userID)
	if err != nil {
		logging.Logger.Error("cant get user URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "cant get user URLs")
	}

	var URLDataSlice []*pb.URLData

	for _, v := range userURL {
		URLData := pb.URLData_builder{
			ShortUrl:    v.ShortURL,
			OriginalUrl: v.OriginalURL,
		}.Build()

		URLDataSlice = append(URLDataSlice, URLData)
	}

	return pb.UserURLsResponse_builder{
		Url: URLDataSlice,
	}.Build(), nil
}
