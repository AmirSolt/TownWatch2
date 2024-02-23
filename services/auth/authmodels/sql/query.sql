

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUsers :many
SELECT * FROM users
WHERE id = ANY($1::text[]);

-- name: CreateUser :one
INSERT INTO users (
    email
) VALUES (
    $1
)
RETURNING *;





-- name: CreateOTP :one
INSERT INTO otps (
    expires_at,
    is_active,
    user_id
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetOTP :one
SELECT * FROM otps
WHERE id = $1 LIMIT 1;

-- name: GetLatestOTPByUser :one
SELECT * FROM otps
WHERE user_id = $1
ORDER BY created_at desc
LIMIT 1;

-- name: DeactivateAllUserOTPs :exec
UPDATE otps
SET is_active = FALSE
WHERE user_id = $1;


-- name: DeactivateOTP :exec
UPDATE otps
SET is_active = FALSE
WHERE id = $1;
