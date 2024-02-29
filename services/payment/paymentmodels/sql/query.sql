
-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1 LIMIT 1;

-- name: GetCustomerByEmail :one
SELECT * FROM customers
WHERE LOWER(email) = LOWER($1) LIMIT 1;

-- name: GetCustomerByUserID :one
SELECT * FROM customers
WHERE user_id = $1 LIMIT 1;

-- name: GetCustomerByStripeCustomerID :one
SELECT * FROM customers
WHERE stripe_customer_id = $1 LIMIT 1;

-- name: UpdateCustomerStripeCustomerID :exec
UPDATE customers
SET stripe_customer_id = $1
WHERE id = $2;

