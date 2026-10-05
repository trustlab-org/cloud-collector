package googleworkspace

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	reports "google.golang.org/api/admin/reports/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const pageSize = 1000

type client struct {
	svc        *reports.Service
	customerID string
}

// newClient authenticates as a service account impersonating an admin via domain-wide delegation.
func newClient(ctx context.Context, endpoint, credentialsFile, subject, customerID string) (*client, error) {
	opts := []option.ClientOption{}
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	if credentialsFile == "" {
		opts = append(opts, option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}))
	} else {
		raw, err := os.ReadFile(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("read credentials: %w", err)
		}
		jwt, err := google.JWTConfigFromJSON(raw, reports.AdminReportsAuditReadonlyScope)
		if err != nil {
			return nil, fmt.Errorf("parse credentials: %w", err)
		}
		jwt.Subject = subject
		httpClient := jwt.Client(context.Background())
		httpClient.Timeout = 60 * time.Second
		opts = append(opts, option.WithHTTPClient(httpClient))
	}
	svc, err := reports.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create reports service: %w", err)
	}
	return &client{svc: svc, customerID: customerID}, nil
}

// page fetches one page of activities; Google returns them newest first.
func (c *client) page(ctx context.Context, app string, start, end time.Time, token string) (*reports.Activities, error) {
	call := c.svc.Activities.List("all", app).
		StartTime(start.UTC().Format(time.RFC3339Nano)).
		EndTime(end.UTC().Format(time.RFC3339Nano)).
		MaxResults(pageSize).
		Context(ctx)
	if c.customerID != "" {
		call = call.CustomerId(c.customerID)
	}
	if token != "" {
		call = call.PageToken(token)
	}
	return call.Do()
}

func retryable(err error) bool {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		switch gerr.Code {
		case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		case http.StatusForbidden:
			for _, e := range gerr.Errors {
				if e.Reason == "rateLimitExceeded" || e.Reason == "userRateLimitExceeded" || e.Reason == "quotaExceeded" {
					return true
				}
			}
		}
		return false
	}
	var nerr net.Error
	if errors.As(err, &nerr) {
		return true
	}
	return !errors.Is(err, context.Canceled)
}
