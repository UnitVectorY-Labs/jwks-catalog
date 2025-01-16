# jwks-catalog

A catalog of JWKS endpoints for popular websites.

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
