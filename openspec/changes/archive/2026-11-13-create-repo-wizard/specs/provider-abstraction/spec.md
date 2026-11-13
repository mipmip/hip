# Spec Delta

## ADDED Requirements

### Requirement: Provider repository creation contract

The provider abstraction SHALL expose creation of a repository for a given
owner, name, description and visibility, returning the created repository as
the provider-agnostic `Repo` model.

#### Scenario: Created repository is returned
- **WHEN** a provider client creates a repository successfully
- **THEN** it returns a `Repo` populated with owner, name, clone URLs, web URL, default branch and visibility

#### Scenario: All providers implement creation
- **WHEN** the default provider registry builds GitHub, GitLab or Codeberg
- **THEN** each client implements repository creation

#### Scenario: Creation failure is returned
- **WHEN** the provider rejects the create request
- **THEN** the client returns an error and no `Repo`
