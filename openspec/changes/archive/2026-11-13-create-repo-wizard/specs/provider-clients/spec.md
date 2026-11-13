# Spec Delta

## ADDED Requirements

### Requirement: GitHub client creates repositories

The GitHub client SHALL create a repository through `POST /user/repos` for the
authenticated user or `POST /orgs/{org}/repos` for an organization, sending the
name, description and requested visibility with provider-side initialization
disabled.

#### Scenario: Create in the self account
- **WHEN** the target owner is the configured GitHub username
- **THEN** the client posts to the authenticated-user endpoint

#### Scenario: Create in an organization
- **WHEN** the target owner is another configured or discovered owner
- **THEN** the client posts to that organization's endpoint

#### Scenario: Visibility is applied
- **WHEN** the user requests a private repository
- **THEN** the request marks it private and the response's visibility is mapped to `private`

### Requirement: GitLab client creates projects

The GitLab client SHALL create a project through `POST /projects` with the
name, path, description and requested visibility, resolving the owner's
namespace id for groups and omitting it for the personal namespace.

#### Scenario: Create in a group
- **WHEN** the target owner is a group or subgroup
- **THEN** the client resolves its namespace id and includes it in the create request

#### Scenario: Create in the personal namespace
- **WHEN** the target owner is the configured GitLab username
- **THEN** the client omits the namespace id

#### Scenario: Provider-side initialization is disabled
- **WHEN** the client creates a project
- **THEN** it requests no provider-created README

### Requirement: Codeberg client creates repositories

The Codeberg client SHALL create a repository through `POST /user/repos` for
the authenticated user or `POST /orgs/{org}/repos` for an organization, with
provider-side initialization disabled.

#### Scenario: Create in the self account
- **WHEN** the target owner is the configured Codeberg username
- **THEN** the client posts to the authenticated-user endpoint

#### Scenario: Create in an organization
- **WHEN** the target owner is another configured or discovered owner
- **THEN** the client posts to that organization's endpoint

### Requirement: Creation errors name the required permission

When a provider rejects a create request for missing authorization, the client
SHALL return an error naming the provider and the credential scope needed to
create the repository.

#### Scenario: Token lacks create permission
- **WHEN** the provider responds with an authorization failure
- **THEN** the error states which scope or permission is required
