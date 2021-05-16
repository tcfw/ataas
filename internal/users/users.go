package users

import (
	"context"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gogo/status"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	usersAPI "pm.tcfw.com.au/source/trader/api/pb/users"
	"pm.tcfw.com.au/source/trader/db"
	"pm.tcfw.com.au/source/trader/internal/passport/utils"
	migrate "pm.tcfw.com.au/source/trader/internal/users/db"
)

var (
	allColumn = []string{
		"id",
		"status",
		"email",
		"firstName",
		"lastname",
		"createdAt",
		"updatedAt",
		"deletedAt",
		"mfa",
		"password",
		"metadata",
		"account",
	}
)

type Server struct {
	usersAPI.UnimplementedUserServiceServer

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

func (s *Server) Me(ctx context.Context, req *usersAPI.Empty) (*usersAPI.User, error) {
	uid, err := utils.UserIDFromContent(ctx)
	if err != nil {
		return nil, err
	}

	q := db.Build().Select(allColumn...).From("users").Where(sq.Eq{"id": uid}).Limit(1)
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()

	user, err := scanUser(res)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Server) Find(ctx context.Context, req *usersAPI.UserRequest) (*usersAPI.User, error) {
	q := db.Build().Select(allColumn...).From("users")

	switch reqType := req.Query.(type) {
	case *usersAPI.UserRequest_Id:
		q = q.Where(sq.Eq{"id": req.GetId()})
	case *usersAPI.UserRequest_Email:
		q = q.Where(sq.Eq{"email": req.GetEmail()})
	default:
		return nil, fmt.Errorf("unknown query type: %s", reqType)
	}

	switch req.Status {
	default:
	case usersAPI.UserRequest_ACTIVE:
		q = q.Where(sq.Eq{"status": usersAPI.User_ACTIVE})
	case usersAPI.UserRequest_DELETED:
		q = q.Where(sq.Eq{"status": usersAPI.User_DELETED})
	case usersAPI.UserRequest_PENDING:
		q = q.Where(sq.Eq{"status": usersAPI.User_PENDING})
	case usersAPI.UserRequest_ANY:
		//filter to not deleted
		q = q.Where(sq.NotEq{"status": usersAPI.UserRequest_DELETED})
	}

	q = q.Limit(1)

	results, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer done()

	if !results.Next() {
		return nil, status.Error(http.StatusNotFound, "User not found")

	}
	user, err := scanUser(results)
	if err != nil {
		return nil, fmt.Errorf("scan in user: %w", err)
	}

	return user, nil
}

func scanUser(res pgx.Row) (*usersAPI.User, error) {
	u := &usersAPI.User{}

	err := res.Scan(
		&u.Id,
		&u.Status,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
		&u.Mfa,
		&u.Password,
		&u.Metadata,
		&u.Account,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}
