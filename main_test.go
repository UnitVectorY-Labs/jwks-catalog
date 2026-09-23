package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProviderEvidence(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "example")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("status.json", `{"oidc":{"status_code":200},"jwks":{"status_code":503}}`)
	write("oidc-headers.json", `{"X-Auth0-L":"[present]","x-auth0-requestid":"[present]"}`)
	write("jwks-headers.json", `{"X-Okta-Request-Id":"[present]"}`)
	service := Service{Id: "example", OpenIDConfiguration: "https://example.com/.well-known/openid-configuration", JWKSURI: "https://example.com/keys"}
	if err := loadProviderEvidence(&service, root); err != nil {
		t.Fatal(err)
	}
	if len(service.Providers) != 1 || service.Providers[0].ID != "auth0" || len(service.ProviderEvidence) != 2 {
		t.Fatalf("unexpected provider evidence: %#v", service)
	}
	for _, evidence := range service.ProviderEvidence {
		if evidence.URL != service.OpenIDConfiguration {
			t.Errorf("stale or wrong endpoint included: %#v", evidence)
		}
	}
	rows := providerEvidenceRows(service.ProviderEvidence)
	if len(rows) != 2 || !rows[0].OIDC || rows[0].OAuth || rows[0].JWKS || !rows[1].OIDC || rows[1].OAuth || rows[1].JWKS {
		t.Fatalf("unexpected evidence matrix: %#v", rows)
	}
}

func TestLoadMultipleProviderEvidence(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "example")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"status.json":       `{"oidc":{"status_code":200},"oauth_authorization_server":{"status_code":200},"jwks":{"status_code":200}}`,
		"oidc-headers.json": `{"X-Sfdc-Edge-Cache":"[present]","X-ForgeRock-TransactionId":"[present]"}`,
		"oauth-authorization-server-headers.json": `{"X-Sfdc-Request-Id":"[present]"}`,
		"jwks-headers.json":                       `{"X-Okta-Request-Id":"[present]","X-Sfdc-Edge-Cache":"[present]"}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	service := Service{Id: "example", OpenIDConfiguration: "https://example.com/oidc", OAuthAuthorizationServerURI: "https://example.com/oauth", JWKSURI: "https://example.com/keys"}
	if err := loadProviderEvidence(&service, root); err != nil {
		t.Fatal(err)
	}
	if len(service.Providers) != 3 || service.Providers[0].ID != "ping" || service.Providers[1].ID != "okta" || service.Providers[2].ID != "salesforce" {
		t.Fatalf("unexpected provider matches: %#v", service.Providers)
	}
	rows := providerEvidenceRows(service.ProviderEvidence)
	if len(rows) != 4 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
	var edge *ProviderEvidenceRow
	for i := range rows {
		if rows[i].Header == "X-Sfdc-Edge-Cache" {
			edge = &rows[i]
		}
	}
	if edge == nil || edge.Provider != "Salesforce" || !edge.OIDC || edge.OAuth || !edge.JWKS {
		t.Fatalf("Salesforce evidence collapsed incorrectly: %#v", edge)
	}
}
