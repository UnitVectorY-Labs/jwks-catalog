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
	Id                  string `yaml:"id"`
	Name                string `yaml:"name"`
	OpenIDConfiguration string `yaml:"openid-configuration"`
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

	// Parse templates
	mainTemplate, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing main template: %v", err)
	}
	snippetTemplate, err := template.ParseFiles("templates/snippet.html")
	if err != nil {
		log.Fatalf("Error parsing snippet template: %v", err)
	}

	// Create output directory
	outputDir := "output"
	servicesDir := filepath.Join(outputDir, "services")
	if err := os.MkdirAll(servicesDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate main index.html
	indexFile := filepath.Join(outputDir, "index.html")
	indexOut, err := os.Create(indexFile)
	if err != nil {
		log.Fatalf("Error creating index file: %v", err)
	}
	defer indexOut.Close()

	if err := mainTemplate.Execute(indexOut, data); err != nil {
		log.Fatalf("Error executing main template: %v", err)
	}

	// Generate services for each service
	for _, service := range data.Services {
		snippetFile := filepath.Join(servicesDir, service.Id+".html")
		snippetOut, err := os.Create(snippetFile)
		if err != nil {
			log.Fatalf("Error creating snippet file for %s: %v", service.Name, err)
		}
		defer snippetOut.Close()

		if err := snippetTemplate.Execute(snippetOut, service); err != nil {
			log.Fatalf("Error executing snippet template for %s: %v", service.Name, err)
		}
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
