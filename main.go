package main

import (
	"bytes"
	"encoding/xml"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

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

func main() {
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
