package broadcast

import (
	"context"
	"log"

	auth "pm.tcfw.com.au/source/trader/internal/passport/utils"
)

func appendContextUserInfo(ctx context.Context, event EventInterface) EventInterface {
	claims, err := auth.TokenClaimsFromContext(ctx)
	if err == nil {
		md := event.GetMetadata()

		if md == nil {
			md = make(map[string]interface{}, 5)
		}

		md["user"] = claims["sub"]
		event.SetMetadata(md)
	} else {
		log.Printf("failed to fetch claims: %s", err)
	}

	return event
}
