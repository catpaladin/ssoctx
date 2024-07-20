package amazon

import (
	"context"
	"strings"
	"testing"

	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/smithy-go/middleware"
)

func TestNewClients(t *testing.T) {
	tests := []struct {
		name               string
		region             string
		withAPIOptionsFunc func(*middleware.Stack) error
		wantErr            bool
	}{
		{
			name:   "ReturnClients",
			region: "us-west-2",
			withAPIOptionsFunc: func(stack *middleware.Stack) error {
				return stack.Finalize.Add(
					middleware.FinalizeMiddlewareFunc(
						"CreateClientsMock",
						func(ctx context.Context, _ middleware.FinalizeInput, _ middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
							operationName := awsMiddleware.GetOperationName(ctx)
							if strings.Contains(operationName, "oidc") {
								return middleware.FinalizeOutput{
									Result: &ssooidc.Client{},
								}, middleware.Metadata{}, nil
							}
							return middleware.FinalizeOutput{
								Result: &sso.Client{},
							}, middleware.Metadata{}, nil
						},
					),
					middleware.Before,
				)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				context.TODO(),
				config.WithRegion(tt.region),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Errorf("NewClients() error = %v", err)
				return
			}
			_, _ = NewClients(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
