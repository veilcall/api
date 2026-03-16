package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dfhgiudhv/privatecall/internal/auth"
	"github.com/dfhgiudhv/privatecall/internal/chat"
	"github.com/dfhgiudhv/privatecall/internal/config"
	"github.com/dfhgiudhv/privatecall/internal/db"
	"github.com/dfhgiudhv/privatecall/internal/expiry"
	"github.com/dfhgiudhv/privatecall/internal/middleware"
	"github.com/dfhgiudhv/privatecall/internal/number"
	"github.com/dfhgiudhv/privatecall/internal/payment"
	rdb "github.com/dfhgiudhv/privatecall/internal/redis"
	"github.com/dfhgiudhv/privatecall/internal/sms"
	"github.com/dfhgiudhv/privatecall/internal/voip"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	redisClient, err := rdb.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisClient.Close()

	// Repositories
	authRepo := auth.NewRepository(pool)
	numRepo := number.NewRepository(pool)
	payRepo := payment.NewRepository(pool)

	// Hubs
	smsHub := sms.NewHub()
	chatHub := chat.NewHub()

	// Services
	authSvc := auth.NewService(authRepo, redisClient, cfg.RecoveryHMACSecret)
	telnyxClient := number.NewTelnyxClient(cfg.TelnyxAPIKey)
	numSvc := number.NewService(numRepo, telnyxClient, smsHub)
	moneroClient := payment.NewMoneroClient(cfg.MoneroRPCURL, cfg.MoneroRPCUser, cfg.MoneroRPCPass)

	prices := map[string]float64{
		"24h": cfg.PlanPrice24HUSD,
		"7d":  cfg.PlanPrice7DUSD,
		"30d": cfg.PlanPrice30DUSD,
	}
	paySvc := payment.NewService(payRepo, moneroClient, prices)
	smsSvc := sms.NewService(cfg.TelnyxAPIKey, numSvc, smsHub)

	// Handlers
	authHandler := auth.NewHandler(authSvc)
	numHandler := number.NewHandler(numSvc)
	payHandler := payment.NewHandler(paySvc)
	reserveHandler := payment.NewReserveHandler(paySvc)
	voipHandler := voip.NewHandler(cfg.FreeSwitchVertoURL, cfg.FreeSwitchVertoSecret)
	chatHandler := chat.NewHandler(chatHub)

	smsHandler, err := sms.NewHandler(smsSvc, smsHub, cfg.TelnyxWebhookSecret)
	if err != nil {
		log.Fatalf("sms handler: %v", err)
	}

	// Background workers
	poller := payment.NewPoller(payRepo, moneroClient, numSvc)
	go poller.Start(ctx)

	expiryWorker := expiry.NewWorker(numRepo, telnyxClient, smsHub)
	go expiryWorker.Start(ctx)

	// Router — gin.New() intentionally: no Logger middleware (avoids IP logging)
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.NoIPLogging()) // first middleware: strip all IP info

	// Public routes
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Telnyx webhook (Telnyx-signed, no user session)
	r.POST("/webhooks/telnyx", smsHandler.TelnyxWebhook)

	// Verto proxy authenticated via token query param
	r.GET("/ws/verto", voipHandler.VertoProxy)

	// Authenticated routes
	authed := r.Group("/")
	authed.Use(middleware.Auth(redisClient))
	{
		authed.POST("/auth/logout", authHandler.Logout)

		authed.GET("/numbers", numHandler.ListNumbers)
		authed.POST("/numbers/reserve", reserveHandler.Reserve)
		authed.DELETE("/numbers/:id", numHandler.ReleaseNumber)

		authed.GET("/payment/:id/status", payHandler.GetStatus)

		authed.POST("/sms/send", smsHandler.SendSMS)

		authed.GET("/voip/token", voipHandler.IssueToken)

		authed.GET("/ws/notify", smsHandler.NotifyWS)
		authed.GET("/ws/chat/:number_id", chatHandler.ChatWS)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx) //nolint:errcheck
}
