//
// Bootstrapping the template service.
//
package service

import (
	"context"

	"github.com/freiheit-com/venus/pkg/logger"
)

// Service implements the RPC methods for the service
type Service struct{}

// NewService instantiates a new service and all of it's dependencies
func NewService() *Service {
	return &Service{}
}

// Shutdown clears all service dependencies
func (s *Service) Shutdown() {
	// Nothing
}

func (s *Service) Echo(ctx context.Context, req *pb.EchoReq) (*pb.EchoRsp, error) {
	logger.Log().Debugf("Echo: Received '%#v'", req.Body)
	return &pb.EchoRsp{Body: req.Body}, nil
}
