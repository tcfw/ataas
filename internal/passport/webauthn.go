package passport

import (
	"fmt"

	"github.com/duo-labs/webauthn/webauthn"
	usersAPI "pm.tcfw.com.au/source/ataas/api/pb/users"
)

func webAuthn() (*webauthn.WebAuthn, error) {
	config := &webauthn.Config{
		RPDisplayName: "Atass",
		RPID:          "atass.io",
		RPOrigin:      "https://staging.atass.io",
	}

	return webauthn.New(config)
}

type webauthnUser struct {
	u *usersAPI.User
}

// User ID according to the Relying Party
func (wau *webauthnUser) WebAuthnID() []byte {
	return []byte(wau.u.Id)
}

// User Name according to the Relying Party
func (wau *webauthnUser) WebAuthnName() string {
	return fmt.Sprintf("%s %s", wau.u.FirstName, wau.u.LastName)
}

// Display Name of the user
func (wau *webauthnUser) WebAuthnDisplayName() string {
	return fmt.Sprintf("%s %s", wau.u.FirstName, wau.u.LastName)
}

// User's icon url
func (wau *webauthnUser) WebAuthnIcon() string {
	return ""
}

// Credentials owned by the user
func (wau *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	fido := wau.u.Mfa.GetFIDO()

	cred := webauthn.Credential{
		ID:              fido.Id,
		PublicKey:       fido.Pk,
		AttestationType: fido.AttestationType,
		Authenticator: webauthn.Authenticator{
			AAGUID:       fido.Authenticator.AAGUID,
			SignCount:    fido.Authenticator.SignCount,
			CloneWarning: fido.Authenticator.CloneWarning,
		},
	}

	return []webauthn.Credential{cred}
}
