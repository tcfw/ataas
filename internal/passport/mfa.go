package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	passportAPI "pm.tcfw.com.au/source/ataas/api/pb/passport"
	usersAPI "pm.tcfw.com.au/source/ataas/api/pb/users"
)

//mfaChallenge signals the auth requestor via a relevant MFA challenge
func (s *Server) mfaChallenge(ctx context.Context, user *usersAPI.User) (*passportAPI.AuthResponse, error) {
	if user.Mfa == nil {
		return nil, nil
	}

	mfa := &passportAPI.MFAResponse{}

	switch mfaType := user.Mfa.MFA.(type) {
	case *usersAPI.MFA_FIDO:
		watn, err := webAuthn()
		if err != nil {
			return nil, err
		}

		webAuthnUser := &webauthnUser{user}

		assert, sessData, err := watn.BeginLogin(webAuthnUser)

		assertData, err := json.Marshal(assert)
		if err != nil {
			return nil, err
		}

		//Store session data for challenge response
		key := fmt.Sprintf("webAuthn.session.%x", assert.Response.AllowedCredentials[0].CredentialID)
		cache, err := s.limiter.cache(ctx)
		if err != nil {
			return nil, err
		}
		err = cache.Set(key, sessData, 5*time.Minute).Err()
		if err != nil {
			return nil, err
		}

		mfa.Type = passportAPI.MFAResponse_FIDO
		mfa.Challenge = &passportAPI.MFAResponse_Fido{Fido: &passportAPI.FIDOChallenge{
			Challenge: string(assertData),
			Timestamp: time.Now().Unix(),
		}}

	case *usersAPI.MFA_TOTP:
		mfa.Type = passportAPI.MFAResponse_TOTP

	case *usersAPI.MFA_SMS:
		mfa.Type = passportAPI.MFAResponse_SMS
		//TODO(tcfw): Send TOTP via SMS to user

	default:
		return nil, fmt.Errorf("Unknown MFA type %v", mfaType)
	}

	return &passportAPI.AuthResponse{Success: false, MFAResponse: mfa}, nil
}

func (s *Server) validateMFA(ctx context.Context, user *usersAPI.User, request *passportAPI.AuthRequest) (bool, error) {
	//TODO(tcfw)
	return false, nil
}
