package basic

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	mdKey         = "authorization"
	mdValuePrefix = "Basic "
)

type Credentials struct {
	UserID   string
	Password string
}

func (c Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	ri, _ := credentials.RequestInfoFromContext(ctx)
	if err := credentials.CheckSecurityLevel(ri.AuthInfo, credentials.PrivacyAndIntegrity); err != nil {
		return nil, fmt.Errorf("unable to transfer basic.Credentials PerRPCCredentials: %v", err)
	}
	return map[string]string{
		mdKey: mdValuePrefix + base64.StdEncoding.EncodeToString([]byte(c.UserID+":"+c.Password)),
	}, nil
}

func (c Credentials) RequireTransportSecurity() bool {
	return true
}

func NewPerRPCCredentials(userID, password string) credentials.PerRPCCredentials {
	return Credentials{UserID: userID, Password: password}
}

func CredentialsFromContext(ctx context.Context) (Credentials, error) {
	var c Credentials
	mv := metadata.ValueFromIncomingContext(ctx, mdKey)
	if mv == nil || len(mv) == 0 {
		return c, fmt.Errorf("basic auth credentials metadata key not found in context")
	}
	decodedCreds, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(mv[0], mdValuePrefix))
	if err != nil {
		return c, err
	}
	c.UserID, c.Password, _ = strings.Cut(string(decodedCreds), ":")
	return c, nil
}