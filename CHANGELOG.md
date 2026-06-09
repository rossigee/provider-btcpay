# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Crossplane runtime upgraded to v2.3.2
- API types migrated from xpv1 to xpv2 (ManagedResourceSpec/Status)
- CI/CD workflows standardized across all providers
- Documentation updated with BTCPay-specific information

### Fixed
- Removed v1 crossplane-runtime indirect dependencies
- Fixed test configuration for Kubernetes secret handling
- Simplified lint configuration using `make lint`

## [0.3.0] - 2026-06-09

### Added
- Release workflow for automated tag-based releases
- Comprehensive security scanning (CodeQL, gosec, govulncheck, TruffleHog)
- Supply chain security scanning (OSV scanner)
- Dependency review with license checking

### Changed
- Updated setup-go from v4 to v6
- Migrated to crossplane-runtime v2.3.2
- Removed disabled publish-artifacts job from CI workflow
- Simplified CI workflow for better maintainability

### Fixed
- Documentation references updated from Plausible to BTCPay
- CLAUDE.md updated with v2 API documentation
- Development guide updated with correct repository URLs

## [0.2.0] - 2026-05-15

### Added
- Store resource management with full CRUD operations
- Invoice resource support for payment processing
- Cross-resource references (Invoice → Store)
- ProviderConfig for authentication and configuration
- Comprehensive unit tests
- GitHub Actions CI/CD pipeline
- Complete documentation and examples

### Features
- **Store Management**: Create, read, update, and delete BTCPay stores
- **Invoice Management**: Create and manage payment invoices
- **Cross-references**: Reference stores from invoices using Kubernetes selectors
- **Status Reporting**: Rich status information with conditions and observations
- **GitOps Support**: Declarative resource management for infrastructure as code

### API Resources
- `Store` (v1alpha1): Manage BTCPay Server stores
- `Invoice` (v1alpha1): Manage payment invoices
- `ProviderConfig` (v1beta1): Configure provider authentication and settings

### Technical Details
- Built on Crossplane provider framework
- Uses BTCPay Greenfield API for programmatic resource management
- Support for custom BTCPay instances via baseURL configuration
- Comprehensive error handling with 404 detection
- Observability through Kubernetes event recording

### Documentation
- Complete README with installation and usage examples
- API reference documentation
- Development setup instructions
- Contributing guidelines

### Testing
- Unit tests for all client operations
- Controller behavior tests
- Mock implementations for testing
- CI pipeline with automated testing

### Build & Deployment
- Docker containerization
- Crossplane package (xpkg) format
- GHCR container registry support
- Automated builds via GitHub Actions

## [0.1.0] - 2026-04-01

### Added
- Initial release of provider-btcpay
- Basic provider structure and scaffolding
- Store and Invoice API definitions
- Provider configuration support

### Notes
- This version was used for initial development and prototyping
- Contains working API definitions with controllers to be implemented
