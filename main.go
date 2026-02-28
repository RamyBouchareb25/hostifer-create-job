package main

import (
	"context"
	"fmt"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

var (
    accountID    = os.Getenv("CF_ACCOUNT_ID")
    tunnelID     = os.Getenv("CF_TUNNEL_ID")
    zoneID       = os.Getenv("CF_ZONE_ID")
    tunnelCNAME  = os.Getenv("CF_TUNNEL_CNAME")
    traefikService = os.Getenv("TRAEFIK_SERVICE")
)

func addSubdomain(subdomain string) error {
    // api, err := cloudflare.NewWithAPIToken(os.Getenv("CF_API_TOKEN"))
    api, err := cloudflare.NewWithAPIToken(os.Getenv("CF_API_TOKEN"))
    if err != nil {
        return err
    }
    ctx := context.Background()

    // 1. Create DNS CNAME record
    zoneRC := cloudflare.ZoneIdentifier(zoneID)
    _, err = api.CreateDNSRecord(ctx, zoneRC, cloudflare.CreateDNSRecordParams{
        Type:    "CNAME",
        Name:    subdomain,           // e.g. "app1" → app1.hostifer.me
        Content: tunnelCNAME,
        Proxied: cloudflare.BoolPtr(true),
        TTL:     1, // auto
    })
    if err != nil {
        return fmt.Errorf("DNS record creation failed: %w", err)
    }

    // 2. Fetch current tunnel config
    accountRC := cloudflare.AccountIdentifier(accountID)
    tunnelConfig, err := api.GetTunnelConfiguration(ctx, accountRC, tunnelID)
    if err != nil {
        return fmt.Errorf("failed to get tunnel config: %w", err)
    }

    // 3. Append new ingress rule (before the catch-all)
    ingress := tunnelConfig.Config.Ingress
    newRule := cloudflare.UnvalidatedIngressRule{
        Hostname: fmt.Sprintf("%s.hostifer.me", subdomain),
        Service:  traefikService,
    }
    // Insert before last catch-all rule
    ingress = append(ingress[:len(ingress)-1], newRule, ingress[len(ingress)-1])

    // 4. Push updated config
    _, err = api.UpdateTunnelConfiguration(ctx, accountRC, cloudflare.TunnelConfigurationParams{
        TunnelID: tunnelID,
        Config:   cloudflare.TunnelConfiguration{Ingress: ingress},
    })
    if err != nil {
        return fmt.Errorf("tunnel config update failed: %w", err)
    }

    return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <subdomain>")
		os.Exit(1)
	}	
	subdomain := os.Args[1]
	if err := addSubdomain(subdomain); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully added subdomain: %s.hostifer.me\n", subdomain)
}
