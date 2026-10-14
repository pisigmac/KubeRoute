# Payments

Stripe integration for task credit purchases.
- Checkout session for one-time top-ups
- Subscription for Pro plan ($49/cluster/month)
- Invoicing for Enterprise
- Webhook: `stripe_webhook_handler` updates `task_credit_history`
