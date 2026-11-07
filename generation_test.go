package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the deployment output, not just individual partial templates: full
// pages must retain no-JS navigation, escaped service data and local runtime
// dependencies that match the source assets byte for byte.
func TestGenerateSelfContainedFullPagesAndFragments(t *testing.T) {
	source, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, dir := range []string{"templates", "assets"} {
		if err := os.CopyFS(filepath.Join(root, dir), os.DirFS(filepath.Join(source, dir))); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	config := `services:
  - id: example
    name: "Example <b>Unsafe</b>"
    openid-configuration: https://example.com/.well-known/openid-configuration
    jwks_uri: https://example.com/keys
    aliases:
      - alias.example.com
`
	if err := os.WriteFile(filepath.Join(root, "data", "services.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	t.Setenv("WEBSITE", "https://catalog.example/")
	t.Setenv("JWKS_OBSERVER_PATH", "")
	originalArgs, originalFlags := os.Args, flag.CommandLine
	t.Cleanup(func() { os.Args, flag.CommandLine = originalArgs, originalFlags })
	os.Args = []string{"jwks-catalog"}
	flag.CommandLine = flag.NewFlagSet("generation-test", flag.ContinueOnError)
	main()
	read := func(name string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(root, "output", name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	for _, asset := range []string{"uvy-tokens.css", "uvy-theme.js", "uvy-elements.css", "vendor/htmx-4.0.0.min.js", "site.js", "licenses/UVY-DESIGN-LICENSE", "licenses/TABLER-ICONS-LICENSE", "licenses/HTMX-LICENSE"} {
		want, e := os.ReadFile(filepath.Join(source, "assets", asset))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(read(asset), want) {
			t.Errorf("generated asset %s differs from the pinned local source", asset)
		}
	}
	nav := read("nav.html")
	for _, page := range []string{"index.html", "service-example.html"} {
		body := read(page)
		if !bytes.Contains(body, nav) {
			t.Errorf("%s lacks initial canonical no-JS navigation", page)
		}
		if !bytes.Contains(body, []byte(`src="./vendor/htmx-4.0.0.min.js"`)) {
			t.Errorf("%s does not load the local HTMX bundle", page)
		}
		if bytes.Contains(body, []byte(`src="https://`)) {
			t.Errorf("%s introduces an external runtime script", page)
		}
		if bytes.Contains(body, []byte("Example <b>Unsafe</b>")) {
			t.Errorf("%s fails to escape service names", page)
		}
	}
	full, snippet := string(read("service-example.html")), string(read("snippets/example.html"))
	if !strings.Contains(full, snippet) {
		t.Error("full service page and HTMX fragment differ")
	}
	if !strings.Contains(full, "Example &lt;b&gt;Unsafe&lt;/b&gt; · JWKS Catalog") {
		t.Error("service document title is missing or not escaped")
	}
	if !strings.Contains(snippet, "https://alias.example.com/.well-known/openid-configuration") {
		t.Error("canonical service aliases were lost")
	}
	for _, name := range []string{"sitemap.xml", "robots.txt", "providers/auth0.svg"} {
		if len(read(name)) == 0 {
			t.Errorf("%s is empty", name)
		}
	}
}
