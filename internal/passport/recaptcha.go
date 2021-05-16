package passport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

const reCAPTCHALink = "https://www.google.com/recaptcha/api/siteverify"

type reCAPTCHARespones struct {
	Success bool `json:"success"`
}

func validateReCAPTCHA(ctx context.Context, token, remoteIP string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	secret := viper.GetString("recpatcha.secret")

	reqVals := url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {remoteIP},
	}

	httpResp, err := http.PostForm(reCAPTCHALink, reqVals)
	if err != nil {
		return false, err
	}

	resp := &reCAPTCHARespones{}

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}
