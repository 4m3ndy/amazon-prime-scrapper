package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"encoding/json"

	"github.com/gocolly/colly/v2"
	//"github.com/gocolly/colly/v2/debug"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
)

type Movie struct {
	Title   	string `json:"title"`
	ReleaseYear string `json:"release_year"`
	Actors		[]string `json:"actors"`
	Poster		string `json:"poster"`
	SimilarIds 	[]string `json:"similar_ids"`
}

// LoginLrConnect returns an access and refresh token upon a successful login ofr the lr connect app
func (m *Movie) LoginLrConnect(ctx context.Context, req *pb.LoginLrConnectRequest) (*pb.LoginLrConnectResponse, error) {
	if len(req.LoginRequest.GetUsername()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty username")
	}

	if len(req.LoginRequest.GetPassword()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty password")
	}

	// ticketGrantingTicket is used as the refresh token
	ticketGrantingTicket, err := sso.GetTicketGrantingTicket(ctx, s.ssoClient, req.LoginRequest.GetUsername(), req.LoginRequest.GetPassword())
	if err != nil {
		logger.Log().WithContext(ctx).WithError(err).Error("failed to get TGT from SSO")
		return nil, pb.HandleError(err, "failed to authenticate user")
	}

	accessToken, refreshToken, err := s.createTokensFromTGT(ctx, ticketGrantingTicket, s.lrConnectTokens)
	if err != nil {
		logger.Log().WithContext(ctx).WithError(err).Error("failed to create tokens")
		return nil, pb.HandleError(err, "failed to authenticate user")
	}

	out := &pb.LoginLrConnectResponse{
		LoginResponse: &pb.LoginResponse{
			RefreshToken: refreshToken,
			AccessToken:  accessToken,
		},
	}

	return out, nil
}
