package auth

import (
	"fmt"
	"net/mail"
	"time"
	"townwatch/base"
	"townwatch/services/auth/authmodels"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const otpExpirationDurationSeconds = 60 * 5  // 5 minutes
const otpRetryExpirationDurationSeconds = 10 // 10 sec

// =========================================================

func (auth *Auth) InitOTP(ctx *gin.Context, email string) *base.ErrorComm {

	user, err := auth.findOrCreateUser(ctx, email)
	if err != nil {
		return err
	}

	otp, err := auth.createOTP(ctx, user)
	if err != nil {
		return err
	}

	errEmail := auth.sendOTPEmail(user, otp)
	if errEmail != nil {
		return errEmail
	}
	return nil
}

func (auth *Auth) DebugOTP(ctx *gin.Context, email string) *base.ErrorComm {

	user, errComm := auth.findOrCreateUser(ctx, email)
	if errComm != nil {
		return errComm
	}
	otp, errComm := auth.createOTP(ctx, user)
	if errComm != nil {
		return errComm
	}

	uuid, err := uuid.FromBytes(otp.ID.Bytes[:])
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("debug error: uuid convserion (%s)", *eventId),
			DevMsg:  err,
		}
	}
	errVOTP := auth.ValidateOTP(ctx, uuid.String())
	if errVOTP != nil {
		return errVOTP
	}

	return nil
}

func (auth *Auth) ValidateOTP(ctx *gin.Context, otpId string) *base.ErrorComm {

	// Find OTP
	id, err := uuid.Parse(otpId)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	otpID := pgtype.UUID{Bytes: id, Valid: true}
	otp, err := auth.Queries.GetOTP(ctx, otpID)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	if !otp.IsActive {
		eventId := sentry.CaptureException(fmt.Errorf("error OTP is not active. OTP: %+v", otp))
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	defer auth.deactivateOTP(ctx, &otp)

	if time.Now().UTC().After(otp.ExpiresAt.Time) {
		err := fmt.Errorf("verification has expired. OTP: %+v | Now: %+v", otp, time.Now().UTC())
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification has expired (%s)", *eventId),
			DevMsg:  err,
		}
	}

	user, err := auth.Queries.GetUser(ctx, otp.UserID)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	errJwt := auth.SetJWTCookie(ctx, &user)
	if errJwt != nil {
		return errJwt
	}

	return nil
}

func (auth *Auth) Signout(ctx *gin.Context) {
	auth.removeJWTCookie(ctx)
}

// =====================================================================

func (auth *Auth) findOrCreateUser(ctx *gin.Context, email string) (*authmodels.User, *base.ErrorComm) {
	var user authmodels.User
	var err error

	_, errEmail := mail.ParseAddress(email)
	if errEmail != nil {
		eventId := sentry.CaptureException(errEmail)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	user, err = auth.Queries.GetUserByEmail(ctx, email)
	if err != nil && err != pgx.ErrNoRows {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
			DevMsg:  err,
		}
	}

	if err == pgx.ErrNoRows {
		user, err = auth.Queries.CreateUser(ctx, email)
		if err != nil {
			eventId := sentry.CaptureException(err)
			return nil, &base.ErrorComm{
				EventID: eventId,
				UserMsg: fmt.Errorf("validation failed (%s)", *eventId),
				DevMsg:  err,
			}

		}
	}

	return &user, nil
}

func (auth *Auth) createOTP(ctx *gin.Context, user *authmodels.User) (*authmodels.Otp, *base.ErrorComm) {

	// =======================
	// make sure last otp happened before otpRetryExpirationDurationSeconds ago
	lastOTP, err := auth.Queries.GetLatestOTPByUser(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification email failed  (%s)", *eventId),
			DevMsg:  err,
		}

	}
	if lastOTP.ID.Valid && time.Now().Add(-time.Second*otpRetryExpirationDurationSeconds).UTC().Before(lastOTP.CreatedAt.Time) {
		err := fmt.Errorf("otp retry timeout. lastOTP: %+v | should be before this time: %+v", lastOTP, time.Now().Add(-time.Second*otpRetryExpirationDurationSeconds).UTC())
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("wait %v seconds before trying again (%s)", otpRetryExpirationDurationSeconds, *eventId),
			DevMsg:  err,
		}
	}
	// =======================

	// =======================
	// make sure all user otp's are inactive before creating a new one
	errDe := auth.Queries.DeactivateAllUserOTPs(ctx, user.ID)
	if errDe != nil {
		eventId := sentry.CaptureException(errDe)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification email failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	// =======================

	otp, err := auth.Queries.CreateOTP(ctx, authmodels.CreateOTPParams{
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second * otpExpirationDurationSeconds).UTC(), Valid: true},
		IsActive:  true,
		UserID:    user.ID,
	})

	if err != nil {
		eventId := sentry.CaptureException(err)
		return nil, &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification email failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return &otp, nil
}

func (auth *Auth) sendOTPEmail(user *authmodels.User, otp *authmodels.Otp) *base.ErrorComm {

	uuid, err := uuid.FromBytes(otp.ID.Bytes[:])
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification email failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	content := "content uuid: " + fmt.Sprintf("%v/join/otp/%v", auth.base.DOMAIN, uuid.String())
	errEmail := auth.base.SendEmail(user.Email, "User", "Town Watch", "Email Verification Link", content)
	if errEmail != nil {
		return errEmail
	}
	return nil
}

func (auth *Auth) deactivateOTP(ctx *gin.Context, otp *authmodels.Otp) *base.ErrorComm {
	err := auth.Queries.DeactivateOTP(ctx, otp.ID)
	if err != nil {
		eventId := sentry.CaptureException(err)
		return &base.ErrorComm{
			EventID: eventId,
			UserMsg: fmt.Errorf("verification email failed (%s)", *eventId),
			DevMsg:  err,
		}
	}
	return nil
}
