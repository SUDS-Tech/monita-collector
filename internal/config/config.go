package config

import "github.com/bastion-framework/bast"

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig `env:"DB"`
	JWT      JWTConfig      `env:"JWT"`
	Stripe   StripeConfig   `env:"STRIPE"`
}

type ServerConfig struct {
	Port int    `env:"PORT" default:"8080"`
	Env  string `env:"APP_ENV" default:"development"`
}

type DatabaseConfig struct {
	URL      string `env:"URL" required:"true"`
	MaxConns int    `env:"MAX_CONNS" default:"25"`
}

type JWTConfig struct {
	Secret string `env:"SECRET" required:"true" secret:"true"`
}

type StripeConfig struct {
	SecretKey     string `env:"SECRET_KEY" secret:"true"`
	WebhookSecret string `env:"WEBHOOK_SECRET" secret:"true"`
	PriceID       string `env:"PRICE_ID"`
}

func Load() (AppConfig, error) {
	return bast.LoadConfig[AppConfig]()
}
