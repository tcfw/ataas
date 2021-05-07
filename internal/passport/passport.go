package passport

import (
	passportAPI "pm.tcfw.com.au/source/trader/api/pb/passport"
)

type Server struct {
	passportAPI.UnimplementedPassportSeviceServer
}
