package excreds

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gogo/status"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	excredsAPI "pm.tcfw.com.au/source/ataas/api/pb/excreds"
	"pm.tcfw.com.au/source/ataas/db"
	migrate "pm.tcfw.com.au/source/ataas/internal/excreds/db"
	passportUtils "pm.tcfw.com.au/source/ataas/internal/passport/utils"
)

const (
	tblName = "excreds"
)

var (
	validEx = map[string]string{
		"binance.com": "binance.com",
	}
)

var (
	allColumns = []string{
		"id",
		"account",
		"exchange",
		"key",
		"secret",
		"createdAt",
	}
)

type Server struct {
	excredsAPI.UnimplementedExCredsServiceServer

	log *logrus.Logger
}

func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		log: logrus.New(),
	}

	err := s.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

func (s *Server) New(ctx context.Context, req *excredsAPI.ExchangeCreds) (*excredsAPI.ExchangeCreds, error) {
	if _, ok := validEx[req.Exchange]; !ok {
		return nil, status.Error(codes.InvalidArgument, "unsupported exchange")
	}

	acn, err := passportUtils.AccountFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Account = acn

	_, err = s.Get(ctx, &excredsAPI.GetRequest{Account: acn, Exchange: req.Exchange})
	if status.Code(err) == codes.OK {
		return nil, status.Error(codes.AlreadyExists, "already exists")
	}

	req.Secret, err = s.encryptSecret(acn, req.Secret)
	if err != nil {
		return nil, err
	}

	q := db.Build().Insert(tblName).Columns("account", "exchange", "key", "secret").Values(
		acn,
		req.Exchange,
		req.Key,
		req.Secret,
	)

	err = db.SimpleExec(ctx, q)
	if err != nil {
		return nil, err
	}

	//read back
	v, err := s.Get(ctx, &excredsAPI.GetRequest{Account: acn, Exchange: req.Exchange})
	v.Secret = ""
	return v, err
}

func (s *Server) List(ctx context.Context, req *excredsAPI.ListRequest) (*excredsAPI.ListResponse, error) {
	acn, err := passportUtils.AccountFromContext(ctx)
	if err != nil {
		return nil, err
	}

	q := db.Build().Select("exchange", "key").From(tblName).Where(sq.Eq{"account": acn})

	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()

	creds := []*excredsAPI.ExchangeCreds{}

	for res.Next() {
		cred := &excredsAPI.ExchangeCreds{}
		err := res.Scan(
			&cred.Exchange,
			&cred.Key,
		)
		if err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}

	return &excredsAPI.ListResponse{Creds: creds}, nil
}

func (s *Server) Delete(ctx context.Context, req *excredsAPI.DeleteRequest) (*excredsAPI.DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (s *Server) Get(ctx context.Context, req *excredsAPI.GetRequest) (*excredsAPI.ExchangeCreds, error) {
	q := db.Build().Select(allColumns...).From(tblName).Where(sq.Eq{"account": req.Account, "exchange": req.Exchange}).Limit(1)
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()

	if !res.Next() {
		return nil, status.Error(codes.NotFound, "not found")
	}

	cred := &excredsAPI.ExchangeCreds{}
	var createdAt time.Time

	err = res.Scan(
		&cred.Id,
		&cred.Account,
		&cred.Exchange,
		&cred.Key,
		&cred.Secret,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	cred.CreatedAt = createdAt.Format(time.RFC3339)

	if req.Decrypt {
		cred.Secret, err = s.decryptSecret(req.Account, cred.Secret)
		if err != nil {
			return nil, err
		}
	}

	return cred, nil
}
