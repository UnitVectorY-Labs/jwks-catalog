package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultKeyHistorySize is the default number of recent keys to show in the history.
	DefaultKeyHistorySize = 10
)

// KeyRecord represents a single key's history for template rendering.
type KeyRecord struct {
	Kid                    string
	Kty                    string
	Use                    string
	Alg                    string
	Crv                    string
	Status                 string // "active" or "inactive"
	DaysActive             int
	FirstObserved          time.Time
	LastObserved           *time.Time // Pointer to handle null for ongoing keys
	FirstObservedFormatted string
	LastObservedFormatted  string
	KeyLength              int // Key length in bits
}

// KeyFile is the JSON structure for a key file from the observer.
type KeyFile struct {
	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Alg string `json:"alg"`
	Crv string `json:"crv,omitempty"`
	Kid string `json:"kid"`

	// Attributes for key history
	FirstObserved time.Time  `json:"first_observed"`
	LastObserved  *time.Time `json:"last_observed,omitempty"`

	// For RSA
	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`

	// For EC
	X string `json:"x,omitempty"`
	Y string `json:"y,omitempty"`
}

// Service represents a JWKS service.
type Service struct {
	Id                  string `yaml:"id"`
	Name                string `yaml:"name"`
	OpenIDConfiguration string `yaml:"openid-configuration"`
	JWKSURI             string `yaml:"jwks_uri"`
}

// ServicePageData holds all the data needed to render a service page.
type ServicePageData struct {
	Service
	ActiveKeys            []KeyRecord
	InactiveKeys          []KeyRecord
	DefaultKeyHistorySize int
}

// Data holds the list of services and content
type Data struct {
	Services []Service     `yaml:"services"`
	Content  template.HTML // Added Content field
}

// SitemapURL represents a single URL entry in sitemap.xml
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap represents the sitemap.xml structure
type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// RobotsTxt represents the robots.txt content
type RobotsTxt struct {
	SitemapURL string
	Disallow   []string
}

// copyFile copies a file from source to destination.
func copyFile(source, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func validateServices(data *Data) error {
	seenIds := make(map[string]bool)
	seenNames := make(map[string]bool)
	seenJWKS := make(map[string]bool)
	seenOIDC := make(map[string]bool)

	for _, service := range data.Services {
		// Check required fields
		if service.Id == "" {
			return fmt.Errorf("service is missing required 'id' field: %v", service)
		}
		if service.Name == "" {
			return fmt.Errorf("service is missing required 'name' field (id: %s)", service.Id)
		}
		if service.JWKSURI == "" {
			return fmt.Errorf("service is missing required 'jwks_uri' field (id: %s)", service.Id)
		}

		// Check for duplicate ID
		if seenIds[service.Id] {
			return fmt.Errorf("duplicate service ID found: %s", service.Id)
		}
		seenIds[service.Id] = true

		// Check for duplicate Name
		if seenNames[service.Name] {
			return fmt.Errorf("duplicate service Name found: %s (id: %s)", service.Name, service.Id)
		}
		seenNames[service.Name] = true

		// Check for duplicate JWKS URI
		if seenJWKS[service.JWKSURI] {
			return fmt.Errorf("duplicate JWKS URI found: %s (service: %s)", service.JWKSURI, service.Id)
		}
		seenJWKS[service.JWKSURI] = true

		// Check for duplicate OpenID Configuration if present
		if service.OpenIDConfiguration != "" {
			if seenOIDC[service.OpenIDConfiguration] {
				return fmt.Errorf("duplicate OpenID Configuration URL found: %s (service: %s)", service.OpenIDConfiguration, service.Id)
			}
			seenOIDC[service.OpenIDConfiguration] = true
		}
	}

	return nil
}

func main() {
	validateFlag := flag.Bool("validate", false, "Validate the services.yaml configuration file")
	flag.Parse()

	// If validate flag is set, only perform validation
	if *validateFlag {
		data, err := loadServices("data/services.yaml")
		if err != nil {
			log.Fatalf("Error loading services: %v", err)
		}

		if err := validateServices(data); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		fmt.Println("Configuration file is valid!")
		os.Exit(0)
	}

	// Retrieve the WEBSITE environment variable
	website := os.Getenv("WEBSITE")
	if website == "" {
		log.Fatal("Environment variable 'WEBSITE' is not set")
	}

	// Ensure the website URL does not have a trailing slash
	if website[len(website)-1] == '/' {
		website = website[:len(website)-1]
	}

	// Load services data
	data, err := loadServices("data/services.yaml")
	if err != nil {
		log.Fatalf("Error loading services: %v", err)
	}

	// Sort services by Id
	sort.Slice(data.Services, func(i, j int) bool {
		return data.Services[i].Id < data.Services[j].Id
	})

	// Parse templates
	mainTemplate, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing main template: %v", err)
	}
	snippetTemplate, err := template.ParseFiles("templates/snippet.html")
	if err != nil {
		log.Fatalf("Error parsing snippet template: %v", err)
	}
	homeTemplate, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Fatalf("Error parsing home template: %v", err)
	}
	navTemplate, err := template.ParseFiles("templates/nav.html")
	if err != nil {
		log.Fatalf("Error parsing nav template: %v", err)
	}

	// Create output directory
	outputDir := "output"
	snippetsDir := filepath.Join(outputDir, "snippets")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Copy style.css to the output directory
	if err := copyFile("assets/style.css", filepath.Join(outputDir, "style.css")); err != nil {
		log.Fatalf("Error copying style.css: %v", err)
	}

	// Generate main index.html with home content
	indexFile := filepath.Join(outputDir, "index.html")
	indexOut, err := os.Create(indexFile)
	if err != nil {
		log.Fatalf("Error creating index file: %v", err)
	}
	defer indexOut.Close()

	contentBuffer := new(bytes.Buffer)
	if err := homeTemplate.Execute(contentBuffer, nil); err != nil {
		log.Fatalf("Error executing home template: %v", err)
	}

	mainData := Data{
		Services: data.Services,
		Content:  template.HTML(contentBuffer.String()), // Assign Content
	}

	if err := mainTemplate.Execute(indexOut, mainData); err != nil {
		log.Fatalf("Error executing main template: %v", err)
	}

	// Generate nav.html
	navFile := filepath.Join(outputDir, "nav.html")
	navOut, err := os.Create(navFile)
	if err != nil {
		log.Fatalf("Error creating nav file: %v", err)
	}
	defer navOut.Close()

	navData := Data{
		Services: data.Services,
	}

	if err := navTemplate.Execute(navOut, navData); err != nil {
		log.Fatalf("Error executing nav template: %v", err)
	}

	// Save home.html to services directory
	homeServiceFile := filepath.Join(snippetsDir, "home.html")
	homeServiceOut, err := os.Create(homeServiceFile)
	if err != nil {
		log.Fatalf("Error creating home service file: %v", err)
	}
	defer homeServiceOut.Close()

	if err := homeTemplate.Execute(homeServiceOut, nil); err != nil {
		log.Fatalf("Error executing home template for services: %v", err)
	}

	// Generate services for each service
	for _, service := range data.Services {
		pageData := ServicePageData{
			Service:               service,
			DefaultKeyHistorySize: DefaultKeyHistorySize,
		}

		// Load key history if observer path is set
		observerPath := os.Getenv("JWKS_OBSERVER_PATH")
		if observerPath != "" {
			activeKeys, inactiveKeys, err := loadKeyHistory(service.Id, observerPath)
			if err != nil {
				log.Printf("Warning: could not load key history for service '%s': %v", service.Id, err)
			} else {
				pageData.ActiveKeys = activeKeys
				pageData.InactiveKeys = inactiveKeys
			}
		}

		// Create the snippet file
		snippetFile := filepath.Join(snippetsDir, service.Id+".html")
		snippetOut, err := os.Create(snippetFile)
		if err != nil {
			log.Fatalf("Error creating service file for %s: %v", service.Name, err)
		}
		defer snippetOut.Close()

		// Capture snippet template output
		snippetBuffer := new(bytes.Buffer)
		if err := snippetTemplate.Execute(snippetBuffer, pageData); err != nil {
			log.Fatalf("Error executing snippet template for %s: %v", service.Name, err)
		}

		if err := snippetTemplate.Execute(snippetOut, pageData); err != nil {
			log.Fatalf("Error executing snippet template for %s: %v", service.Name, err)
		}

		// Create service file
		serviceFile := filepath.Join(outputDir, "service-"+service.Id+".html")
		serviceOut, err := os.Create(serviceFile)
		if err != nil {
			log.Fatalf("Error creating service file for %s: %v", service.Name, err)
		}
		defer serviceOut.Close()

		serviceData := Data{
			Services: data.Services,
			Content:  template.HTML(snippetBuffer.String()), // Assign snippet content
		}

		if err := mainTemplate.Execute(serviceOut, serviceData); err != nil {
			log.Fatalf("Error executing snippet template for %s: %v", service.Name, err)
		}
	}

	// Generate sitemap.xml
	if err := generateSitemap(outputDir, website); err != nil {
		log.Fatalf("Error generating sitemap.xml: %v", err)
	}

	// Generate robots.txt
	if err := generateRobotsTxt(outputDir, website); err != nil {
		log.Fatalf("Error generating robots.txt: %v", err)
	}

	log.Println("Static files generated successfully!")
}

func loadServices(filename string) (*Data, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data Data
	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func generateSitemap(outputDir, website string) error {
	var sitemap Sitemap
	sitemap.Xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

	// Collect all .html files in outputDir excluding /snippets
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip files in /snippets
		relPath, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}
		if filepath.Dir(relPath) == "snippets" {
			return nil
		}

		// Include only .html files
		if filepath.Ext(info.Name()) == ".html" {
			url := SitemapURL{
				Loc:     website + "/" + relPath,
				LastMod: info.ModTime().Format("2006-01-02"),
			}
			// Special cases for index.html
			if info.Name() == "index.html" {
				url.Loc = website + "/"
			}
			sitemap.URLs = append(sitemap.URLs, url)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Sort URLs alphabetically
	sort.Slice(sitemap.URLs, func(i, j int) bool {
		return sitemap.URLs[i].Loc < sitemap.URLs[j].Loc
	})

	// Create sitemap.xml file
	sitemapFile := filepath.Join(outputDir, "sitemap.xml")
	sitemapOut, err := os.Create(sitemapFile)
	if err != nil {
		return err
	}
	defer sitemapOut.Close()

	// Marshal sitemap to XML with indentation
	xmlData, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return err
	}

	// Add XML header
	finalSitemap := []byte(xml.Header + string(xmlData))
	if _, err := sitemapOut.Write(finalSitemap); err != nil {
		return err
	}

	log.Println("sitemap.xml generated successfully.")
	return nil
}

func generateRobotsTxt(outputDir, website string) error {
	robots := RobotsTxt{
		SitemapURL: website + "/sitemap.xml",
		Disallow:   []string{"/snippets"},
	}

	// Parse robots.txt template from file
	tmpl, err := template.ParseFiles("templates/robots.txt")
	if err != nil {
		return err
	}

	// Create robots.txt file
	robotsFile := filepath.Join(outputDir, "robots.txt")
	robotsOut, err := os.Create(robotsFile)
	if err != nil {
		return err
	}
	defer robotsOut.Close()

	// Execute template
	if err := tmpl.Execute(robotsOut, robots); err != nil {
		return err
	}

	log.Println("robots.txt generated successfully.")
	return nil
}

// loadKeyHistory loads, processes, and sorts the key history for a service.
func loadKeyHistory(serviceID, observerPath string) ([]KeyRecord, []KeyRecord, error) {
	serviceDataPath := filepath.Join(observerPath, "data", serviceID)
	if _, err := os.Stat(serviceDataPath); os.IsNotExist(err) {
		return nil, nil, nil // No observer data for this service, skip silently.
	}

	keysPath := filepath.Join(serviceDataPath, "keys")
	keyFiles, err := os.ReadDir(keysPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read keys directory: %w", err)
	}

	var allKeys []KeyRecord
	for _, kf := range keyFiles {
		if filepath.Ext(kf.Name()) != ".json" {
			continue
		}

		keyFilePath := filepath.Join(keysPath, kf.Name())
		keyData, err := os.ReadFile(keyFilePath)
		if err != nil {
			log.Printf("Warning: could not read key file %s: %v", keyFilePath, err)
			continue
		}

		var keyFile KeyFile
		if err := json.Unmarshal(keyData, &keyFile); err != nil {
			log.Printf("Warning: could not parse key file %s: %v", keyFilePath, err)
			continue
		}

		status := "inactive"
		if keyFile.LastObserved == nil {
			status = "active"
		}

		lastObservedFormatted := ""
		if keyFile.LastObserved != nil {
			lastObservedFormatted = keyFile.LastObserved.Format("2006-01-02")
		}

		allKeys = append(allKeys, KeyRecord{
			Kid:                    keyFile.Kid,
			Kty:                    keyFile.Kty,
			Use:                    keyFile.Use,
			Alg:                    keyFile.Alg,
			Crv:                    keyFile.Crv,
			Status:                 status,
			FirstObserved:          keyFile.FirstObserved,
			LastObserved:           keyFile.LastObserved,
			DaysActive:             calculateDaysActive(keyFile.FirstObserved, keyFile.LastObserved),
			FirstObservedFormatted: keyFile.FirstObserved.Format("2006-01-02"),
			LastObservedFormatted:  lastObservedFormatted,
			KeyLength:              calculateKeyLength(&keyFile),
		})
	}

	var activeKeys, inactiveKeys []KeyRecord
	for _, key := range allKeys {
		if key.Status == "active" {
			activeKeys = append(activeKeys, key)
		} else {
			inactiveKeys = append(inactiveKeys, key)
		}
	}

	// Sort active keys by FirstObserved (newest first)
	sort.Slice(activeKeys, func(i, j int) bool {
		return activeKeys[i].FirstObserved.After(activeKeys[j].FirstObserved)
	})

	// Sort inactive keys by LastObserved (newest first)
	sort.Slice(inactiveKeys, func(i, j int) bool {
		if inactiveKeys[i].LastObserved == nil || inactiveKeys[j].LastObserved == nil {
			return false // Should not happen for inactive keys
		}
		return (*inactiveKeys[i].LastObserved).After(*inactiveKeys[j].LastObserved)
	})

	// Limit inactive keys to the most recent N
	if len(inactiveKeys) > DefaultKeyHistorySize {
		inactiveKeys = inactiveKeys[:DefaultKeyHistorySize]
	}

	return activeKeys, inactiveKeys, nil
}

// calculateDaysActive computes the number of days a key was active.
func calculateDaysActive(start time.Time, end *time.Time) int {
	until := time.Now()
	if end != nil {
		until = *end
	}
	duration := until.Sub(start)
	return int(duration.Hours() / 24)
}

// calculateKeyLength determines the key length in bits for RSA and EC keys.
func calculateKeyLength(kf *KeyFile) int {
	switch kf.Kty {
	case "RSA":
		// N is base64url-encoded big-endian integer
		if kf.N == "" {
			return 0
		}
		nBytes, err := decodeBase64URL(kf.N)
		if err != nil {
			return 0
		}
		return len(nBytes) * 8
	case "EC":
		// Use curve name
		switch kf.Crv {
		case "P-256":
			return 256
		case "P-384":
			return 384
		case "P-521":
			return 521
		}
		return 0
	default:
		return 0
	}
}

// decodeBase64URL decodes a base64url-encoded string.
func decodeBase64URL(s string) ([]byte, error) {
	// base64.RawURLEncoding ignores padding
	return base64.RawURLEncoding.DecodeString(s)
}
