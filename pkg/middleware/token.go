package middleware

import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/slaghuis/YC-Image/pkg/config"
)

type AccessClaims struct {
    Sub     string `json:"sub"` // user id
    Role    string `json:"role"`
    Email   string `json:"email,omitempty"`
    Name    string `json:"name,omitempty"`
    jwt.RegisteredClaims
}

func parseAccessToken(cfg config.JwtConfigurations, token string) (*AccessClaims, error) {
    t, err := jwt.ParseWithClaims(token, &AccessClaims{}, func(tok *jwt.Token) (interface{}, error) {
        return []byte(cfg.JWTSigningKey), nil
    })
    if err != nil || !t.Valid {
        return nil, err
    }
    claims, _ := t.Claims.(*AccessClaims)
    return claims, nil
}
