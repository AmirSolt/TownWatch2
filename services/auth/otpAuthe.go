package auth

import (
	"fmt"
	"time"
	"townwatch/services/auth/authmodels"

	"github.com/gin-gonic/gin"
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

	auth.ValidateOTP(ctx, string(otp.ID.Bytes[:]))

	return nil
}

func (auth *Auth) ResendOTP(ctx *gin.Context, email string) error {

	// find user
	user, err := auth.Queries.GetUserByEmail(ctx, email)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("error user email lookup: %w", err)
	}

	// =======================
	// make sure last otp happened before otpRetryExpirationDurationSeconds ago
	lastOTP, err := auth.Queries.GetLatestOTPByUser(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("error latest otp lookup: %w", err)
	}
	if time.Now().Add(-time.Second * otpRetryExpirationDurationSeconds).UTC().Before(lastOTP.CreatedAt.Time) {
		return fmt.Errorf("you have to wait %v after sending OTP: %w", otpRetryExpirationDurationSeconds, err)
	}
	// =======================

	otp, err := auth.createOTP(ctx, &user)
	if err != nil {
		return err
	}

	errEmail := auth.sendOTPEmail(&user, otp)
	if errEmail != nil {
		return errEmail
	}

	return nil
}

func (auth *Auth) ValidateOTP(ctx *gin.Context, otpId string) error {

	// Find OTP
	otp, err := auth.Queries.GetOTP(ctx, pgtype.UUID{Bytes: stringToByte16(otpId), Valid: true})
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

	lastOTP, err := auth.Queries.GetLatestOTPByUser(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("error latest otp lookup: %w", err)
	}

	if lastOTP.ID != otp.ID {
		return fmt.Errorf("otp does not match latest user otp: %w", err)
	}

	auth.SetJWTCookie(ctx, &user)

	return nil
}

func Signout(ctx *gin.Context) {
	removeJWTCookie(ctx)
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
	otp, err := auth.Queries.CreateOTP(ctx, authmodels.CreateOTPParams{
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second * otpExpirationDurationSeconds).UTC(), Valid: true},
		IsActive:  true,
		UserID:    user.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("error OTP creation: %w", err)
	}
	return &otp, nil
}

func (auth *Auth) sendOTPEmail(user *authmodels.User, otp *authmodels.Otp) error {
	content := "content" + string(otp.ID.Bytes[:])
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
