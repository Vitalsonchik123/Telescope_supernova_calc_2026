package redis

import (
	"context"
	"time"
)

const jwtPrefix = "jwt."

// getJWTKey — префикс + токен, чтобы ключи не пересекались с другими
func getJWTKey(token string) string {
	return jwtPrefix + token
}

// WriteJWTToBlacklist — кладёт токен в blacklist на время его жизни
func (c *Client) WriteJWTToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return c.client.Set(ctx, getJWTKey(token), true, ttl).Err()
}

// CheckJWTInBlacklist — возвращает true, если токен в blacklist
func (c *Client) CheckJWTInBlacklist(ctx context.Context, token string) bool {
	exists, err := c.client.Exists(ctx, getJWTKey(token)).Result()
	if err != nil {
		return false
	}
	return exists > 0
}
