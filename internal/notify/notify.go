package notify

import (
	"context"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"
	notifyAPI "pm.tcfw.com.au/source/ataas/api/pb/notify"
	usersAPI "pm.tcfw.com.au/source/ataas/api/pb/users"
)

var domain string = "mg.tcfw.com.au"
var privateAPIKey string = ""

type Server struct {
	notifyAPI.UnimplementedNotifyServiceServer

	log *logrus.Logger
}

func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		log: logrus.New(),
	}

	return s, nil
}

func (s *Server) Send(ctx context.Context, req *notifyAPI.SendRequest) (*notifyAPI.SendResponse, error) {
	mg := mailgun.NewMailgun(domain, privateAPIKey)

	users, err := usersSvc()
	if err != nil {
		s.log.Warn(err)
		return nil, err
	}

	user, err := users.Find(ctx, &usersAPI.UserRequest{
		Query:  &usersAPI.UserRequest_Id{Id: req.Uid},
		Status: usersAPI.UserRequest_ACTIVE,
	})
	if err != nil {
		s.log.Warn(err)
		return nil, err
	}

	message := mg.NewMessage("ataas@tcfw.com.au", req.Title, req.Body, user.Email)

	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		s.log.Warn(err)
	}

	s.log.Infof("Email Sent - ID: %s Resp: %s\n", id, resp)

	return &notifyAPI.SendResponse{}, nil
}
