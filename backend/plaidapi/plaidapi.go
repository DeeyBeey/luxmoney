// Package plaidapi wraps the Plaid SDK for the rest of the backend.
package plaidapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	plaid "github.com/plaid/plaid-go/v51"
	"github.com/plaid/plaid-go/v51/client"
	"github.com/plaid/plaid-go/v51/option"
)

// Client is the backend's handle to Plaid. Create one with New and share it.
type Client struct {
	sdk *client.Client
}

// New builds a Client for the given Plaid environment ("sandbox" or "production").
func New(clientID, secret, env string) *Client {
	// http.Client's zero value has no timeout, so a hung call to Plaid
	// would block the request forever. Cap every Plaid call instead.
	httpClient := &http.Client{Timeout: 15 * time.Second}

	return &Client{
		sdk: client.NewClient(
			option.WithClientID(new(clientID)),
			option.WithSecret(secret),
			option.WithBaseURL(baseURL(env)),
			option.WithHTTPClient(httpClient),
		),
	}
}

// sandboxInstitutionID is "First Platypus Bank", Plaid's fake test bank.
const sandboxInstitutionID = "ins_109508"

// CreateSandboxPublicToken asks Plaid's sandbox for a public token for its
// test bank, skipping the Link UI. It only works against sandbox.
func (c *Client) CreateSandboxPublicToken(ctx context.Context) (string, error) {
	resp, err := c.sdk.Sandbox.CreatePublicToken(ctx, &plaid.SandboxPublicTokenCreateRequest{
		InstitutionID:   sandboxInstitutionID,
		InitialProducts: []plaid.Products{plaid.ProductsTransactions},
	})
	if err != nil {
		return "", fmt.Errorf("sandbox public token: %w", err)
	}
	return resp.PublicToken, nil
}

// Item is one linked bank login. AccessToken is secret and never leaves the backend.
type Item struct {
	AccessToken string
	ItemID      string
}

// ExchangePublicToken trades a one-time public token for the Item's
// permanent access token and ID. Store the result: it can't be fetched again.
func (c *Client) ExchangePublicToken(ctx context.Context, publicToken string) (Item, error) {
	resp, err := c.sdk.Item.ExchangePublicToken(ctx, &plaid.ItemPublicTokenExchangeRequest{
		PublicToken: publicToken,
	})
	if err != nil {
		return Item{}, fmt.Errorf("exchange public token: %w", err)
	}
	return Item{AccessToken: resp.AccessToken, ItemID: resp.ItemID}, nil
}

// baseURL maps PLAID_ENV to Plaid's API host.
// loadConfig already rejected anything other than these two values.
func baseURL(env string) string {
	if env == "production" {
		return plaid.Environments.Production
	}
	return plaid.Environments.Sandbox
}
