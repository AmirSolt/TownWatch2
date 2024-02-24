
-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1 LIMIT 1;

-- name: GetCustomerByStripeCustomerID :one
SELECT * FROM customers
WHERE stripe_customer_id = $1 LIMIT 1;

-- name: UpdateCustomerStripeCustomerID :exec
UPDATE customers
SET stripe_customer_id = $1
WHERE id = $2;

-- name: UpdateCustomerSubAndTier :exec
UPDATE customers
SET 
stripe_subscription_id = $1,
tier_id = $2
WHERE id = $3;
