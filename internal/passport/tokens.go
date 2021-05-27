package passport

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	passportAPI "pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/api/pb/users"
	"pm.tcfw.com.au/source/ataas/db"
)

func init() {
	viper.SetDefault("passport.token.key", "passport.key")
	viper.SetDefault("passport.token.cert", "passport.cert")
}

func (s *Server) verifyToken(ctx context.Context, request *passportAPI.VerifyTokenRequest) (*passportAPI.VerifyTokenResponse, error) {
	if request.Token == "" {
		return nil, fmt.Errorf("Invalid token format")
	}

	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is ECDSA
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return GetKeyPublic()
	})

	if err != nil || !token.Valid {
		peer, _ := peer.FromContext(ctx)
		log.Printf("Invalid Token Used by '%s' due to '%s'", peer.Addr, err)
		return &passportAPI.VerifyTokenResponse{
			Valid: false,
		}, nil
	}

	claims, _ := token.Claims.(jwt.MapClaims)

	isRevoked, err := s.isTokenRevoked(ctx, claims)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if isRevoked {
		return &passportAPI.VerifyTokenResponse{
			Valid:   false,
			Revoked: true,
		}, nil
	}

	return &passportAPI.VerifyTokenResponse{
		Valid:       token.Valid,
		TokenExpire: int64(claims["exp"].(float64)),
	}, nil
}

//UserClaims takes in a user and applies the standard user claims
func UserClaims(user *users.User) map[string]interface{} {
	claims := make(map[string]interface{}, 2)

	claims["sub"] = user.Id
	claims["acn"] = user.Account

	return claims
}

//GetKeyPrivate reads a private PEM formatted RSA cert
func GetKeyPrivate() (*ecdsa.PrivateKey, error) {
	dat, err := ioutil.ReadFile(viper.GetString("passport.token.key"))
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(dat)
	if err != nil {
		log.Printf("Err: %s", err)
		return nil, err
	}

	return key, nil
}

//GetKeyPublic reads a public PEM formatted RSA cert
func GetKeyPublic() (*ecdsa.PublicKey, error) {
	dat, err := ioutil.ReadFile(viper.GetString("passport.token.cert"))
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseECPublicKeyFromPEM(dat)
	if err != nil {
		log.Printf("Err: %s", err)
		return nil, err
	}

	return key, nil
}

//makeNewToken creates a new JWT token for the specific user
func (s *Server) makeNewToken(ctx context.Context, extraClaims map[string]interface{}) (string, *jwt.Token, error) {
	signer := jwt.New(jwt.SigningMethodES256)

	//set claims
	claims := make(jwt.MapClaims)
	hostname, err := os.Hostname()
	if err != nil {
		claims["iss"] = "passport.ataas.io"
	} else {
		claims["iss"] = hostname + ".ataas.io"
	}

	expr := time.Now().UTC().Add(time.Hour * 168)

	claims["nbf"] = time.Now().UTC().Unix() - 1
	claims["iat"] = time.Now().UTC().Unix()
	claims["exp"] = expr.Unix() // 1 week
	claims["jti"] = uuid.New().String()

	for claimKey, claimValue := range extraClaims {
		claims[claimKey] = claimValue
	}

	signer.Claims = claims

	key, err := GetKeyPrivate()
	if err != nil {
		return "", nil, fmt.Errorf("error extracting the key")
	}
	signer.Header["kid"] = "03c56b083c89ade910b54300e09a3afe"

	tokenString, err := signer.SignedString(key)
	if err != nil {
		return "", nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nil, fmt.Errorf("failed to parse metadata")
	}

	ip := ""
	ua := ""

	if len(md.Get("x-forwarded-for")) > 0 {
		ip = md.Get("x-forwarded-for")[0]
	}
	if len(md.Get("grpcgateway-user-agent")) > 0 {
		ua = md.Get("grpcgateway-user-agent")[0]
	}

	q := db.Build().Insert("sessions").Columns("jti", "sub", "exp", "ip", "ua").Values(
		claims["jti"].(string),
		claims["sub"].(string),
		expr,
		ip,
		ua,
	)

	err = db.SimpleExec(ctx, q)
	if err != nil {
		return "", nil, err
	}

	return tokenString, signer, nil
}

/*MakeTestToken makes a temporary JWT token which expires in 4
 *seconds specifically for testing purposes
 */
func MakeTestToken(user *users.User) (string, error) {
	signer := jwt.New(jwt.SigningMethodHS256)

	claims := make(jwt.MapClaims)
	claims["iss"] = "testing.ataas.io"
	claims["nbf"] = time.Now().Unix() - 1
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Second * 4).Unix()
	claims["jti"] = uuid.New().String()

	extraClaims := UserClaims(user)

	for claimKey, claimValue := range extraClaims {
		claims[claimKey] = claimValue
	}

	signer.Claims = claims
	return signer.SignedString([]byte("this is a super secret key, DO NOT USE FOR PRODUCTION"))
}

//MakeNewRefresh creates a new refresh token
func MakeNewRefresh(ctx context.Context) *string {
	randomString := RandomString(256)
	return &randomString
}

//RandomString generates a n lengthed string (cryptographically)
func RandomString(n int) string {
	var randomBytes = make([]byte, n/2)
	rand.Read(randomBytes)

	return hex.EncodeToString(randomBytes)
}
