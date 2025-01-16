package main

import (
	"bytes"
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

// Data holds the list of services and content
type Data struct {
	Services []Service     `yaml:"services"`
	Content  template.HTML // Added Content field
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
	homeTemplate, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Fatalf("Error parsing home template: %v", err)
	}

	// Create output directory
	outputDir := "output"
	snippetsDir := filepath.Join(outputDir, "snippets")
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
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

		// Create the snippet file
		snippetFile := filepath.Join(snippetsDir, service.Id+".html")
		snippetOut, err := os.Create(snippetFile)
		if err != nil {
			log.Fatalf("Error creating service file for %s: %v", service.Name, err)
		}
		defer snippetOut.Close()

		// Capture snippet template output
		snippetBuffer := new(bytes.Buffer)
		if err := snippetTemplate.Execute(snippetBuffer, service); err != nil {
			log.Fatalf("Error executing snippet template for %s: %v", service.Name, err)
		}

		if err := snippetTemplate.Execute(snippetOut, service); err != nil {
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
