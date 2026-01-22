package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor handles JWT validation
func AuthInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 1. Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata missing")
		}

		// 2. Get the Authorization header
		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token missing")
		}

		// 3. Parse and Validate JWT
		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			// Validate signing method to prevent algorithm confusion attacks
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 4. Proceed to the actual handler
		return handler(ctx, req)
	}
}
