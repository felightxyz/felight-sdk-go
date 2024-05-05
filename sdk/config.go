package client

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ErrorNilConfig            = errors.New("config is nil")
	ErrEndpointNotProvided    = errors.New("API endpoint must be provided")
	ErrAccessTokenNotProvided = errors.New("access token must be provided")
)

type Config struct {
	Url             string
	AccessToken     string
	WriteBufferSize int
	ReadBufferSize  int
	ConnWindowSize  int32
	WindowSize      int32
	// max attempts to re-subscribe streams when they are broken
	MaxRetry int
	// in seconds
	RetryInterval int64

	// configures a connection level security credentials, TLS is enabled with local CA by default if not provided
	transportOption grpc.DialOption
}

func NewConfig() *Config {
	return &Config{
		WriteBufferSize: 1024 * 8,
		ReadBufferSize:  1024 * 8,
		ConnWindowSize:  1024 * 512,
		WindowSize:      1024 * 256,

		MaxRetry:      64,
		RetryInterval: 2,
	}
}

func ValidateConfig(c *Config) error {
	if c == nil {
		return ErrorNilConfig
	}
	if c.Url == "" {
		return ErrEndpointNotProvided
	}
	if c.AccessToken == "" {
		return ErrAccessTokenNotProvided
	}
	return nil
}

func (c *Config) SetTransportOption(creds credentials.TransportCredentials) {
	c.transportOption = grpc.WithTransportCredentials(creds)
}

func (c *Config) defaultDialOptions() []grpc.DialOption {
	config := `{
		"methodConfig": [{
			"name": [{"service": "API"}],
			"retryPolicy": {
				"MaxAttempts": 20,
				"InitialBackoff": "2s",
				"MaxBackoff": "60s",
				"BackoffMultiplier": 1.2,
				"RetryableStatusCodes": ["UNAVAILABLE", "ABORTED"]
			}
		}]
	}`
	return []grpc.DialOption{grpc.WithDefaultServiceConfig(config)}
}

// DialOptions convert configuration into grpc dail options
func (c *Config) DialOptions() []grpc.DialOption {
	opts := c.defaultDialOptions()

	if c.transportOption == nil {
		c.transportOption = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}
	opts = append(opts, c.transportOption)
	opts = append(opts, grpc.WithWriteBufferSize(c.WriteBufferSize))
	opts = append(opts, grpc.WithReadBufferSize(c.ReadBufferSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(c.ConnWindowSize))
	opts = append(opts, grpc.WithInitialWindowSize(c.WindowSize))

	opts = append(opts, grpc.WithPerRPCCredentials(authCredentials{accessToken: c.AccessToken}))
	return opts
}

type authCredentials struct {
	accessToken string
}

func (ac authCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		authHeaderKey: ac.accessToken,
	}, nil
}

func (ac authCredentials) RequireTransportSecurity() bool {
	return false
}
