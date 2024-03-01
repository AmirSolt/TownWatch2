Notes:
- Currently you can not add folders to domains/**/templates
- Every time you add a domain, it must have a registery.go file and be registered to the app within main.go

Stripe:
- Make sure to limit customers to one subscription: https://docs.stripe.com/payments/checkout/limit-subscriptions
- All other configurations are managed within the app, go to services/payment/main.go for more info