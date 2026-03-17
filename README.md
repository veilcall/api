# api

Go backend for Veilcall — anonymous virtual number service.

## Endpoints

```
POST   /auth/register       Register anonymously — returns one-time recovery code
POST   /auth/login          Login with recovery code — returns session token
POST   /auth/logout

POST   /numbers/reserve     Create Monero payment for a number
GET    /payment/:id/status  Poll payment confirmation
GET    /numbers             List active numbers
DELETE /numbers/:id         Release number early

POST   /sms/send            Send SMS from a number
POST   /webhooks/telnyx     Inbound SMS webhook (Ed25519 verified)

WS     /ws/notify           Real-time SMS and expiry notifications
WS     /ws/verto            FreeSWITCH Verto WebRTC proxy
WS     /ws/chat/:number_id  E2E encrypted chat relay
```

## Privacy

- No IP addresses stored (`gin.New()`, not `gin.Default()`)
- No CDR (call detail records)
- No SMS content persisted
- Sessions in Redis memory-only (no disk write)
- Recovery codes stored as HMAC-SHA256 only

## Stack

Go 1.23 · Gin · pgx/v5 · go-redis/v9 · gorilla/websocket
