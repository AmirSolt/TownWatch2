package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"townwatch/base"
	"townwatch/services/auth/authmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (auth *Auth) ValidateUser(ctx *gin.Context) (*authmodels.User, *base.ErrorComm) {
	// get it from cookie
	tokenString, err := ctx.Cookie("Authorization")
	if tokenString == "" || err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("user validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	// parse and validate token
	jwt, errComm := dencryptJWT([]byte(tokenString), auth.base.JWE_SECRET_KEY)
	if errComm != nil {
		return nil, errComm
	}

	// find user and check exp
	user, errComm := auth.ValidateJWTByUser(ctx, jwt)
	if errComm != nil {
		return nil, errComm
	}

	return user, nil
}

func (auth *Auth) SetJWTCookie(ctx *gin.Context, user *authmodels.User) *base.ErrorComm {

	uuid, err := uuid.FromBytes(user.ID.Bytes[:])
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	jwt := JWT{
		ID:  uuid.String(),
		IP:  ctx.ClientIP(),
		EXP: time.Now().Add(time.Second * jwtExpirationDurationSeconds).Unix(),
	}

	jwtEncrypted, errComm := encryptJWT(jwt, auth.base.JWE_SECRET_KEY)
	if errComm != nil {
		return errComm
	}

	// attach to cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"Authorization",
		string(jwtEncrypted),
		jwtExpirationDurationSeconds,
		"/",
		auth.base.DOMAIN,
		true,
		true,
	)
	return nil
}

func (auth *Auth) ValidateJWTByUser(ctx *gin.Context, jwt *JWT) (*authmodels.User, *base.ErrorComm) {

	if jwt.EXP < time.Now().Unix() {
		err := fmt.Errorf("jwt expired. JWT.EXP: %v | NOW: %+v", jwt.EXP, time.Now().Unix())
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	if jwt.IP != ctx.ClientIP() {
		err := fmt.Errorf("jwt IP mismatch. JWT.IP: %v | ctx.ClientIP(): %v", jwt.IP, ctx.ClientIP())
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	id, err := uuid.Parse(jwt.ID)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	userID := pgtype.UUID{Bytes: id, Valid: true}
	user, err := auth.Queries.GetUser(ctx, userID)
	if err != nil && err != pgx.ErrNoRows {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &user, nil
}

func (auth *Auth) removeJWTCookie(ctx *gin.Context) {

	ctx.SetCookie(
		"Authorization",
		"",
		jwtExpirationDurationSeconds,
		"/",
		auth.base.DOMAIN,
		true,
		true,
	)

}

// ==============================================================

func encryptJWT(jwt JWT, jwe_secret_key string) ([]byte, *base.ErrorComm) {

	payload, err := json.Marshal(jwt)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	encrypted, err := jwe.Encrypt(payload, jwe.WithKey(jwa.A128GCMKW, []byte(jwe_secret_key)))
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return encrypted, nil
}
func dencryptJWT(encryptedJWT []byte, jwe_secret_key string) (*JWT, *base.ErrorComm) {
	decrypted, err := jwe.Decrypt(encryptedJWT, jwe.WithKey(jwa.A128GCMKW, []byte(jwe_secret_key)))
	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	var jwt JWT
	errJson := json.Unmarshal(decrypted, &jwt)
	if errJson != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("authorization failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	return &jwt, nil
}
