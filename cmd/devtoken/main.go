// Command devtoken mints a short-lived HS256 JWT for local development and
// demos. Production tokens come from your identity provider, not this tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"netorchestrator/internal/platform"
)

func main() {
	secret := flag.String("secret", os.Getenv("GATEWAY_JWT_SECRET"), "HS256 secret (>=32 bytes)")
	sub := flag.String("sub", "", "subject (user id)")
	ttl := flag.Duration("ttl", time.Hour, "token lifetime")
	flag.Parse()
	if len(*secret) < 32 || *sub == "" {
		fmt.Fprintln(os.Stderr, "usage: devtoken -secret <32+ bytes> -sub <user> [-ttl 1h]")
		os.Exit(2)
	}
	fmt.Println(platform.SignHS256([]byte(*secret), platform.Claims{Subject: *sub, Expires: time.Now().Add(*ttl).Unix()}))
}
