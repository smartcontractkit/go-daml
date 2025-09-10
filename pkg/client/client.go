package client

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/noders-team/go-daml/pkg/auth"
)

type Client struct {
	config *Config
	conn   *grpc.ClientConn
}

func NewClient(config *Config) *Client {
	return &Client{
		config: config,
	}
}

func (c *Client) Connect(ctx context.Context) (*Connection, error) {
	opts := c.buildDialOptions()

	conn, err := grpc.DialContext(ctx, c.config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DAML ledger: %w", err)
	}

	c.conn = conn
	return NewConnection(c, conn), nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) buildDialOptions() []grpc.DialOption {
	var opts []grpc.DialOption

	if c.config.TLS != nil {
		tlsConfig := c.buildTLSConfig()
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))

		if c.config.Auth != nil {
			bearerAuth := c.createBearerAuth()
			opts = append(opts, grpc.WithPerRPCCredentials(bearerAuth))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if c.config.Auth != nil {
			bearerAuth := c.createBearerAuth()
			opts = append(opts,
				grpc.WithUnaryInterceptor(bearerAuth.UnaryInterceptor()),
				grpc.WithStreamInterceptor(bearerAuth.StreamInterceptor()),
			)
		}
	}

	return opts
}

func (c *Client) buildTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		ServerName:         c.config.TLS.ServerName,
		InsecureSkipVerify: c.config.TLS.InsecureSkipVerify,
	}

	if c.config.TLS.CertFile != "" {
	}

	return tlsConfig
}

func (c *Client) createBearerAuth() *auth.BearerTokenAuth {
	if c.config.Auth.TokenProvider != nil {
		return auth.NewBearerTokenProvider(c.config.Auth.TokenProvider)
	}
	return auth.NewBearerToken(c.config.Auth.Token)
}

type Connection struct {
	client *Client
	conn   *grpc.ClientConn
}

func NewConnection(client *Client, conn *grpc.ClientConn) *Connection {
	return &Connection{
		client: client,
		conn:   conn,
	}
}

func (c *Connection) GRPCConn() *grpc.ClientConn {
	return c.conn
}
