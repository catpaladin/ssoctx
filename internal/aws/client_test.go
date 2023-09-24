// Package aws contains all the aws logic
package aws

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

func TestCreateClients(t *testing.T) {
	type args struct {
		ctx    context.Context
		region string
	}
	tests := []struct {
		name  string
		args  args
		want  *ssooidc.Client
		want1 *sso.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CreateClients(tt.args.ctx, tt.args.region)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateClients() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CreateClients() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
