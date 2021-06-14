package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/codes"
	"pm.tcfw.com.au/source/ataas/internal/utils/tracing"
)

const reCAPTCHALink = "https://www.google.com/recaptcha/api/siteverify"

type reCAPTCHARespones struct {
	Success     bool     `json:"success"`
	Score       float32  `json:"score"`
	ChallengeTS string   `json:"challenge_ts"`
	Action      string   `json:"action"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func validateReCAPTCHA(ctx context.Context, token, remoteIP string) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "recaptch_verify")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	secret := viper.GetString("recaptcha.secret")

	reqVals := url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {remoteIP},
	}

	span.AddEvent(fmt.Sprintf("%+v", reqVals))

	httpResp, err := http.PostForm(reCAPTCHALink, reqVals)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	resp := &reCAPTCHARespones{}

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	span.AddEvent(fmt.Sprintf("%+v", resp))
	return resp.Success, nil
}
