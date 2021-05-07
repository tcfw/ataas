package users

import (
	usersAPI "pm.tcfw.com.au/source/trader/api/pb/users"
)

type Server struct {
	usersAPI.UnimplementedUserServiceServer
}
