package common

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/notify-plugin/pkg/apis"
)

var ErrCodeMap = make(map[error]codes.Code)

func RegisterErr(originErr error, errCode codes.Code) {
	ErrCodeMap[originErr] = errCode
}

func ConvertErr(err error) error {
	if err == nil {
		return nil
	}
	if code, ok := ErrCodeMap[errors.Cause(err)]; ok {
		return status.Error(code, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func init() {
	RegisterErr(ErrConfigMiss, codes.FailedPrecondition)
}

type Server struct {
	Sender ISender
}

func NewServer(sender ISender) *Server {
	return &Server{Sender:sender}
}

func (s *Server) Send(ctx context.Context, req *apis.SendParams) (*apis.Empty, error) {
	empty := &apis.Empty{}
	if !s.Sender.IsReady(ctx) {
		err := status.Error(codes.FailedPrecondition, NOTINIT)
		return empty, err
	}
	log.Debugf("recevie msg, contact: %s, title: %s, content: %s", req.Contact, req.Title, req.Message)
	err := s.Sender.Send(ctx, req)
	return empty, ConvertErr(err)
}

func (s *Server) UpdateConfig(ctx context.Context, req *apis.UpdateConfigParams) (empty *apis.Empty, err error) {
	empty = new(apis.Empty)
	defer func() {
		if err != nil {
			log.Errorf("update config error: %s", err.Error())
		}
	}()
	if req.Configs == nil {
		return empty, status.Error(codes.InvalidArgument, ConfigNil)
	}
	log.Debugf("update configs: %v", req.Configs)
	err = s.Sender.UpdateConfig(ctx, req.Configs)
	return empty, ConvertErr(err)
}

func (s *Server) ValidateConfig(ctx context.Context, req *apis.UpdateConfigParams) (*apis.ValidateConfigReply, error) {
	if req.Configs == nil {
		return nil, status.Error(codes.InvalidArgument, ConfigNil)
	}
	log.Debugf("validate configs: %v", req.Configs)
	formatConfig, err := s.Sender.CheckConfig(ctx, req.Configs)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	rep, err := s.Sender.ValidateConfig(ctx, formatConfig)
	if err != nil {
		return nil, ConvertErr(err)
	}
	return rep, nil
}

func (s *Server) UseridByMobile(ctx context.Context, req *apis.UseridByMobileParams) (*apis.UseridByMobileReply, error) {
	if !s.Sender.IsReady(ctx) {
		return nil, status.Error(codes.FailedPrecondition, NOTINIT)
	}
	log.Debugf("fetch userid by mobile %s", req.Mobile)
	userId, err := s.Sender.FetchContact(ctx, req.Mobile)
	if err != nil {
		return nil, ConvertErr(err)
	}
	return &apis.UseridByMobileReply{
		Userid:               userId,
	}, nil
}
