package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"streamgogambler/internal/adapters/config"
	"streamgogambler/internal/adapters/gui"
	"streamgogambler/internal/adapters/healthcheck"
	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/adapters/storage"
	"streamgogambler/internal/adapters/twitch"
	"streamgogambler/internal/application"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	if !AcquireSingleInstanceLock() {
		log.Printf("[ERROR] Another instance of StreamGoGambler is already running.")
		gui.ShowErrorDialog("StreamGoGambler is already running", "Another instance of the application is already running. Only one instance can run at a time.")
		os.Exit(1)
	}
	defer ReleaseSingleInstanceLock()

	envPath := config.ResolveEnvPath()
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("[WARN] Could not load .env from %s: %v", envPath, err)
	}

	missingVars := config.GetMissingVariables()
	if len(missingVars) > 0 {
		log.Printf("[INFO] Missing configuration variables: %v", missingVars)
		log.Printf("[INFO] Launching setup dialog...")

		defaults := config.GetDefaultValues()
		result := gui.ShowSetupDialog(missingVars, defaults)

		if !result.Completed {
			log.Printf("[INFO] Setup canceled by user. Exiting.")
			ReleaseSingleInstanceLock()
			os.Exit(0) //nolint:gocritic
		}

		if err := config.SaveConfigToEnv(envPath, result.Values); err != nil {
			log.Fatalf("[FATAL] Failed to save configuration: %v", err)
		}

		log.Printf("[INFO] Configuration saved to %s", envPath)
	}

	cfgStore, err := config.NewEnvStore(envPath)
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	cfg := cfgStore.GetConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.NewFromString(cfg.LogLevel)
	logger.Infof(ctx, "StreamGoGambler %s (commit: %s, built: %s)", version, commit, buildDate)

	chatClient := twitch.NewClient(
		cfg.Username,
		cfgStore.GetOAuth(),
		twitch.WithBucketSize(cfg.SayBucketSize),
		twitch.WithRefillMs(cfg.SayRefillMs),
		twitch.WithLogger(logger),
	)

	trustedUsersPath := storage.ResolveTrustedUsersPath(envPath)
	trustedStore := storage.NewTrustedUsersStore(trustedUsersPath)

	botService := application.NewBotService(cfgStore, chatClient, logger, trustedStore)

	if cfg.HealthPort > 0 {
		healthServer := healthcheck.NewHealthServer(cfg.HealthPort, botService, logger)
		if err := healthServer.Start(ctx); err != nil {
			logger.Errorf(ctx, "Failed to start health server: %v", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- botService.Start(ctx)
	}()

	if cfg.GUIEnabled {
		HideConsole()

		botGUI := gui.New(botService, cfg.MaxLogsLines)

		logger.SetCallback(func(message string) {
			botGUI.AppendLog(message)
		})

		go func() {
			select {
			case sig := <-sigChan:
				logger.Infof(ctx, "Received signal %v, shutting down...", sig)
			case err := <-errChan:
				if err != nil {
					logger.Errorf(ctx, "Bot error: %v", err)
				}
			}

			cancel()
			botService.Stop()
			botGUI.Stop()
			logger.Infof(ctx, "Bot stopped")
		}()

		botGUI.Run()

		cancel()
		botService.Stop()
		logger.Infof(ctx, "Application terminated")
	} else {
		logger.Infof(ctx, "Running in headless mode (GUI disabled)")
		select {
		case sig := <-sigChan:
			logger.Infof(ctx, "Received signal %v, shutting down...", sig)
		case err := <-errChan:
			if err != nil {
				logger.Errorf(ctx, "Bot error: %v", err)
			}
		}

		cancel()
		botService.Stop()
		logger.Infof(ctx, "Application terminated")
	}
}
