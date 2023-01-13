package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type jwtUser struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type tokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type claims struct {
	jwt.RegisteredClaims
}

func (j *auth) GenerateTokenPair(user *jwtUser) (tokenPairs, error) {
	// Create a token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	//  Get token.claims, then cast it to a jwt.MapClaims object
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.Id) // Subject
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix() // Unix timestamp
	claims["typ"] = "JWT"

	// Set the expiry for JWT
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// Create signed JWT token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return tokenPairs{}, err
	}

	// Create a refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.Id)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// Set expiry for refresh
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// Create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return tokenPairs{}, err
	}

	// Create tokenPairs
	var tokenPairs = tokenPairs{
		Token: signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
		
	return tokenPairs, nil
}