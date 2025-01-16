package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Service represents a JWKS service
type Service struct {
	Name                string `yaml:"name"`
	OpenIDConfiguration string `yaml:"openid_configuration"`
	JWKSURI             string `yaml:"jwks_uri"`
}

// Data holds the list of services
type Data struct {
	Services []Service `yaml:"services"`
}

func main() {
	// Load services data
	data, err := loadServices("data/services.yaml")
	if err != nil {
		log.Fatalf("Error loading services: %v", err)
	}

	// Parse template
	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Create output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate static file
	outputFile := filepath.Join(outputDir, "index.html")
	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	log.Printf("Static file generated at %s", outputFile)
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
