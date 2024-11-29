package dns

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflareDNSProvider struct {
	api *cloudflare.API
}

// NewCloudflareDNS creates a new CloudflareDNS instance using the provided API token
func NewCloudflareDNS(apiToken string) (*CloudflareDNSProvider, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, err
	}

	return &CloudflareDNSProvider{
		api: api,
	}, nil
}

func (c *CloudflareDNSProvider) AddRecord(zoneID, name, recordType, content string, proxied bool) error {
	record := cloudflare.CreateDNSRecordParams{
		Name:    name,
		Type:    recordType,
		Content: content,
		Proxied: &proxied,
	}

	_, err := c.api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), record)
	return err
}

func (c *CloudflareDNSProvider) GetRecord(zoneID, name string) (*cloudflare.DNSRecord, error) {
	record, err := c.api.GetDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), name)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *CloudflareDNSProvider) AllRecords(zoneID string) ([]cloudflare.DNSRecord, error) {
	records, _, err := c.api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (c *CloudflareDNSProvider) DeleteRecord(zoneID, recordID string) error {
	return c.api.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), recordID)
}
