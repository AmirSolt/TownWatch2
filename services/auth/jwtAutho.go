package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"townwatch/services/auth/authmodels"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
)

const jwtExpirationDurationSeconds = 60 * 60 * 24 * 15 // 15 days

type JWT struct {
	ID  string `json:"id"`
	IP  string `json:"ip"`
	EXP int64  `json:"exp"`
}

func (auth *Auth) ValidateUser(ctx *gin.Context) (*authmodels.User, error) {
	// get it from cookie
	tokenString, err := ctx.Cookie("Authorization")
	if err != nil {
		return nil, fmt.Errorf("jwt not found on cookie: %w", err)
	}

	// parse and validate token
	jwt, err := auth.ParseJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("jwt parse failed: %w", err)
	}

	// find user and check exp
	user, err := auth.ValidateJWTByUser(ctx, jwt)
	if err != nil {
		return nil, fmt.Errorf("jwt validation by user failed: %w", err)
	}

	return user, nil
}

func (auth *Auth) SetJWTCookie(ctx *gin.Context, user *authmodels.User) error {

	uuid, err := uuid.FromBytes(user.ID.Bytes[:])
	if err != nil {
		return err
	}

	jwt := JWT{
		ID:  uuid.String(),
		IP:  ctx.ClientIP(),
		EXP: time.Now().Add(time.Second * jwtExpirationDurationSeconds).Unix(),
	}

	jwtEncrypted, err := encryptJWT(jwt, auth.base.JWE_SECRET_KEY)
	if err != nil {
		return fmt.Errorf("jwt authorization failed: %w", err)
	}

	// attach to cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"Authorization",
		string(jwtEncrypted),
		jwtExpirationDurationSeconds,
		"/",
		"",
		true,
		true,
	)
	return nil
}

func removeJWTCookie(ctx *gin.Context) {

	ctx.SetCookie(
		"Authorization",
		"",
		jwtExpirationDurationSeconds,
		"/",
		"",
		true,
		true,
	)

}

func (auth *Auth) ParseJWT(jwtEncrypted string) (*JWT, error) {
	jwt, err := dencryptJWT([]byte(jwtEncrypted), auth.base.JWE_SECRET_KEY)
	if err != nil {
		return nil, fmt.Errorf("jwt decryption failed: %w", err)
	}
	return jwt, nil
}

func (auth *Auth) ValidateJWTByUser(ctx *gin.Context, jwt *JWT) (*authmodels.User, error) {

	if jwt.EXP < time.Now().Unix() {
		return nil, fmt.Errorf("error jwt is expired")
	}

	if jwt.IP != ctx.ClientIP() {
		return nil, fmt.Errorf("error jwt is from an invalid IP")
	}

	id, err := uuid.Parse(jwt.ID)
	if err != nil {
		return nil, err
	}
	userID := pgtype.UUID{Bytes: id, Valid: true}
	user, err := auth.Queries.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error jwt user not found")
	}

	return &user, nil
}

// ==============================================================

func encryptJWT(jwt JWT, jwe_secret_key string) ([]byte, error) {

	payload, err := json.Marshal(jwt)
	if err != nil {
		return nil, fmt.Errorf("failed to json marshal payload: %w", err)
	}
	encrypted, err := jwe.Encrypt(payload, jwe.WithKey(jwa.A128GCMKW, []byte(jwe_secret_key)))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt jwt payload: %w", err)
	}
	return encrypted, nil
}
func dencryptJWT(encryptedJWT []byte, jwe_secret_key string) (*JWT, error) {
	decrypted, err := jwe.Decrypt(encryptedJWT, jwe.WithKey(jwa.A128GCMKW, []byte(jwe_secret_key)))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var jwt JWT
	errJson := json.Unmarshal(decrypted, &jwt)
	if errJson != nil {
		return nil, fmt.Errorf("failed to json unmarshal payload: %w", err)
	}

	return &jwt, nil
}
