# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.0] - 2025-10-27

### Added
- **PaymentMethod Resource**: Full CRUD operations for store payment methods
- **User Resource**: User management with role-based access control
- **Webhook Resource**: Webhook configuration for event notifications
- **v1beta1 API Versions**: New stable API versions for Invoice and Store resources
- **Namespace-scoped Resources**: Support for namespaced deployments with cross-namespace references

### Changed
- **API Stability**: Promoted Invoice and Store to v1beta1 with backward compatibility
- **Controller Architecture**: Refactored controllers to support multiple API versions
- **Build System**: Updated Makefile and CI workflows for improved reliability

### Fixed
- **Documentation**: Corrected references from "Plausible" to "BTCPay" throughout codebase
- **Generated Code**: Regenerated managed resource files for consistency

### Technical Improvements
- Enhanced error handling across all controllers
- Improved test coverage for new resources
- Updated dependencies and build tooling

## [v0.2.1] - 2025-08-15

### Added
- Enhanced Invoice resource with additional metadata fields
- Improved error handling for BTCPay API rate limits
- Additional unit tests for edge cases

### Fixed
- Store creation race conditions
- Invoice status synchronization issues

## [v0.2.0] - 2025-08-01

### Added
- **Invoice Resource**: Full invoice lifecycle management
- **Cross-Resource References**: Invoice to Store referencing
- **Payment Tracking**: Real-time invoice status updates
- Expanded test suite with integration tests

### Changed
- ProviderConfig API to v1beta1 for stability
- Enhanced BTCPay client with additional endpoints

## [v0.1.0] - 2025-07-07

### Added
- Initial release of BTCPay Server Crossplane Provider
- **Store Resource**: Full CRUD operations for BTCPay stores
- **Provider Configuration**: Secure API key management via Kubernetes secrets
- Comprehensive unit tests and CI/CD pipeline
- Complete documentation and examples

### Features
- **Store Management**: Create, configure, and manage BTCPay Server stores
- **Multi-tenant Support**: Team-based store organization
- **Status Reporting**: Rich conditions and observations
- **Security**: Kubernetes-native secret management

### API Resources
- `Store` (v1alpha1): Manage BTCPay Server stores
- `ProviderConfig` (v1beta1): Configure provider authentication

### Technical Details
- Built on Crossplane provider framework
- Full BTCPay Greenfield API integration
- Comprehensive error handling and observability
- Docker containerization and Crossplane package support