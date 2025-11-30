[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active) [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/jwks-catalog)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/jwks-catalog)

# jwks-catalog

A catalog of JWKS endpoints for popular websites.

Available at: [https://jwks-catalog.unitvectorylabs.com/](https://jwks-catalog.unitvectorylabs.com/)

## Overview

[JSON Web Key Sets](https://datatracker.ietf.org/doc/html/rfc7517) (JWKS) are a standard mechanism used in modern authentication systems to facilitate secure communication and validation of digital signatures. A JWKS URL provides a publicly accessible endpoint that hosts cryptographic keys used by identity providers to sign tokens like JSON Web Tokens (JWTs).

This catalog aggregates JWKS URLs from popular services such as Google, GitHub, Microsoft, Apple, and others creating a resource for developers to quickly find and reference JWKS endpoints.

## Contributing

This catalog is open to contributions which can be added by adding entries to the following file: [services.yaml](https://github.com/UnitVectorY-Labs/jwks-catalog/blob/main/data/services.yaml)

Each entry in the YAML file should contain the following fields:

- `id`: A unique identifier for the service
- `name`: The name of the service
- `openid-configuration`: The OpenID configuration URL for the service (optional)
- `jwks_uri`: The JWKS URL for the service

## Site Generation

The JWKS Catalog utilizes lightweight tech stack for static page generation. This process converts the data defined in a YAML file into the static HTML website, hosted on GitHub Pages.

Tech Stack**:

- **Go**: The primary programming language used for the transformation of YAML data into structured HTML files.
- **HTMX**: Enhances the interactivity of the generated site by enabling dynamic content loading without full-page reloads. HTMX is used to fetch and display service-specific JWKS details without fully reloading the page.
- **HTML Templates**: Custom templates are used to define the layout and structure of the site. Go’s templating engine facilitates the seamless integration of dynamic data into the pre-defined templates.

The workflow includes:

1. Parsing the [services.yaml](https://github.com/UnitVectorY-Labs/jwks-catalog/blob/main/data/services.yaml) which serves as the primary data source.
2. Using Go templates to generate the primary `index.html` and complete page for each service.  
3. Generating “snippet” pages for each service to support HTMX-driven dynamic content loading without requiring a full-page reload.

The final static files are deployed to GitHub Pages using the [jwks-catalog-go-pages-deploy.yml](https://github.com/UnitVectorY-Labs/jwks-catalog/blob/main/.github/workflows/jwks-catalog-go-pages-deploy.yml) GitHub Action workflow, which is automatically triggered on every push to the `main` branch.
