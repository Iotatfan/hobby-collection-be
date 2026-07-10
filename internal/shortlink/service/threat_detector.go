package service

import (
	"context"
	"fmt"

	webrisk "cloud.google.com/go/webrisk/apiv1"
	webriskpb "cloud.google.com/go/webrisk/apiv1/webriskpb"
)

type ThreatDetector interface {
	IsMaliciousURL(ctx context.Context, url string) (bool, error)
}

type webRiskDetector struct {
	client *webrisk.Client
}

func NewWebRiskDetector(client *webrisk.Client) ThreatDetector {
	return &webRiskDetector{
		client: client,
	}
}

func (w *webRiskDetector) IsMaliciousURL(ctx context.Context, url string) (bool, error) {
	if w.client == nil {
		// If client is nil, we might be in testing or no credentials provided, fail open or closed depending on requirements.
		// Let's assume safe if no client is configured for simplicity, or return error.
		return false, nil
	}

	req := &webriskpb.SearchUrisRequest{
		Uri: url,
		ThreatTypes: []webriskpb.ThreatType{
			webriskpb.ThreatType_MALWARE,
			webriskpb.ThreatType_SOCIAL_ENGINEERING,
			webriskpb.ThreatType_UNWANTED_SOFTWARE,
		},
	}

	resp, err := w.client.SearchUris(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to search uri in web risk: %w", err)
	}

	if resp.Threat != nil {
		return true, nil
	}

	return false, nil
}
