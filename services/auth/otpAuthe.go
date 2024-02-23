package auth

import (
	"fmt"
	"time"
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

func (auth *Auth) InitOTP(ctx *gin.Context, email string) error {

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

func (auth *Auth) DebugOTP(ctx *gin.Context, email string) error {

	user, err := auth.findOrCreateUser(ctx, email)
	if err != nil {
		return err
	}
	otp, err := auth.createOTP(ctx, user)
	if err != nil {

		return err
	}

	uuid, err := uuid.FromBytes(otp.ID.Bytes[:])
	if err != nil {
		return err
	}
	errVOTP := auth.ValidateOTP(ctx, uuid.String())
	if errVOTP != nil {
		return errVOTP
	}

	return nil
}

func (auth *Auth) ValidateOTP(ctx *gin.Context, otpId string) error {

	// Find OTP
	id, err := uuid.Parse(otpId)
	if err != nil {
		return err
	}
	otpID := pgtype.UUID{Bytes: id, Valid: true}
	otp, err := auth.Queries.GetOTP(ctx, otpID)
	if err != nil {
		return fmt.Errorf("error OTP lookup: %w", err)
	}

	if !otp.IsActive {
		return fmt.Errorf("error OTP is not active: %w", err)
	}
	defer auth.deactivateOTP(ctx, &otp)

	if time.Now().UTC().After(otp.ExpiresAt.Time) {
		return fmt.Errorf("error OTP is expired: %w", err)
	}

	user, err := auth.Queries.GetUser(ctx, otp.UserID)
	if err != nil {
		return fmt.Errorf("error user not found by OTP: %w", err)
	}

	errJwt := auth.SetJWTCookie(ctx, &user)
	if errJwt != nil {
		return fmt.Errorf("error Setting JWT Cookie: %w", errJwt)
	}

	return nil
}

func (auth *Auth) Signout(ctx *gin.Context) {
	auth.removeJWTCookie(ctx)
}

// =====================================================================

func (auth *Auth) findOrCreateUser(ctx *gin.Context, email string) (*authmodels.User, error) {
	var user authmodels.User
	var err error

	user, err = auth.Queries.GetUserByEmail(ctx, email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error user email lookup: %w", err)
	}

	if err == pgx.ErrNoRows {
		user, err = auth.Queries.CreateUser(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("error user creation: %w", err)
		}
	}

	return &user, nil
}

func (auth *Auth) createOTP(ctx *gin.Context, user *authmodels.User) (*authmodels.Otp, error) {

	// =======================
	// make sure last otp happened before otpRetryExpirationDurationSeconds ago
	lastOTP, err := auth.Queries.GetLatestOTPByUser(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error latest otp lookup: %w", err)
	}
	if lastOTP.ID.Valid && time.Now().Add(-time.Second*otpRetryExpirationDurationSeconds).UTC().Before(lastOTP.CreatedAt.Time) {
		return nil, fmt.Errorf("you have to wait %v after sending OTP: %w", otpRetryExpirationDurationSeconds, err)
	}
	// =======================

	// =======================
	// make sure all user otp's are inactive before creating a new one
	errDe := auth.Queries.DeactivateAllUserOTPs(ctx, user.ID)
	if errDe != nil {
		return nil, fmt.Errorf("error DeactivateAllUserOTPs: %w", errDe)
	}
	// =======================

	otp, err := auth.Queries.CreateOTP(ctx, authmodels.CreateOTPParams{
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second * otpExpirationDurationSeconds).UTC(), Valid: true},
		IsActive:  true,
		UserID:    user.ID,
	})

	eventId := sentry.CaptureException(fmt.Errorf("error OTP creation: %w", err))
	if err != nil {

		return nil, fmt.Errorf("error OTP creation: %v", eventId)
	}
	return &otp, nil
}

func (auth *Auth) sendOTPEmail(user *authmodels.User, otp *authmodels.Otp) error {

	uuid, err := uuid.FromBytes(otp.ID.Bytes[:])
	if err != nil {
		return err
	}
	content := "content uuid: " + fmt.Sprintf("%v/join/otp/%v", auth.base.DOMAIN, uuid.String())
	errEmail := auth.base.SendEmail(user.Email, "User", "Town Watch", "Email Verification Link", content)
	if errEmail != nil {
		return fmt.Errorf("error OTP email could not be sent: %w", errEmail)
	}
	return nil
}

func (auth *Auth) deactivateOTP(ctx *gin.Context, otp *authmodels.Otp) error {
	err := auth.Queries.DeactivateOTP(ctx, otp.ID)
	if err != nil {
		return fmt.Errorf("deactivating otp failed: %w", err)
	}
	return nil
}
