package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port string

	DBDSN    string
	RedisURL string

	TelnyxAPIKey        string
	TelnyxSIPUsername   string
	TelnyxSIPPassword   string
	TelnyxWebhookSecret string

	MoneroRPCURL  string
	MoneroRPCUser string
	MoneroRPCPass string

	FreeSwitchVertoURL    string
	FreeSwitchVertoSecret string

	RecoveryHMACSecret string

	PlanPrice24HUSD float64
	PlanPrice7DUSD  float64
	PlanPrice30DUSD float64

	GinMode string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),

		DBDSN:    mustEnv("DB_DSN"),
		RedisURL: mustEnv("REDIS_URL"),

		TelnyxAPIKey:        mustEnv("TELNYX_API_KEY"),
		TelnyxSIPUsername:   mustEnv("TELNYX_SIP_USERNAME"),
		TelnyxSIPPassword:   mustEnv("TELNYX_SIP_PASSWORD"),
		TelnyxWebhookSecret: mustEnv("TELNYX_WEBHOOK_SECRET"),

		MoneroRPCURL:  getEnv("MONERO_RPC_URL", "http://monero:18082/json_rpc"),
		MoneroRPCUser: mustEnv("MONERO_RPC_USER"),
		MoneroRPCPass: mustEnv("MONERO_RPC_PASS"),

		FreeSwitchVertoURL:    getEnv("FREESWITCH_VERTO_URL", "ws://freeswitch:8081"),
		FreeSwitchVertoSecret: mustEnv("FREESWITCH_VERTO_SECRET"),

		RecoveryHMACSecret: mustEnv("RECOVERY_HMAC_SECRET"),

		PlanPrice24HUSD: parseFloat(getEnv("PLAN_PRICE_24H_USD", "2.99")),
		PlanPrice7DUSD:  parseFloat(getEnv("PLAN_PRICE_7D_USD", "7.99")),
		PlanPrice30DUSD: parseFloat(getEnv("PLAN_PRICE_30D_USD", "19.99")),

		GinMode: getEnv("GIN_MODE", "release"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env var not set: " + key)
	}
	return v
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
