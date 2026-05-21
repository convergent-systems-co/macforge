# GOALS.md

## Project Name

MacForge

## Mission

MacForge is civilization-grade Apple release infrastructure for macOS software.

MacForge provides deterministic, auditable, repeatable Apple signing, certificate lifecycle management, CI secret management, packaging, notarization, verification, and publishing for macOS software distributed outside the App Store.

Goal:

identity → certificate → keychain → signing → packaging → notarization → verification → publish

Eliminate:

- Apple signing complexity
- Keychain ambiguity
- CI credential pain
- Release pipeline vendor lock-in
- Paywalled Apple release tooling

## Principles

- Deterministic
- Auditable
- Reproducible
- Local private key ownership
- Dedicated keychain isolation
- CI/CD native
- Open source first
- Enterprise capable
- Civilization-grade reliability

## Core Features

### Identity

- Dedicated keychain creation
- Private key generation
- CSR generation
- Certificate import
- Certificate validation
- Certificate rotation
- Certificate expiration monitoring
- Multi-team support

### Signing

- Developer ID Application support
- Hardened Runtime
- Timestamping
- Entitlement support
- Artifact signing validation

### Packaging

- zip
- dmg
- pkg
- app bundle support

### Notarization

- Apple notarytool integration
- Submission
- Wait
- Status
- Stapling
- Verification

### Verification

- codesign verification
- spctl verification
- Gatekeeper validation
- Artifact attestations

### CI Integration

- GitHub Actions
- GitLab
- Azure DevOps

### GitHub Action

macforge-action

Responsibilities:

- Keychain import
- Signing
- Packaging
- Notarization
- Verification
- Publish artifacts

### Release Command

macforge release

Performs:

build → sign → package → notarize → verify → publish

## Architecture

CLI

- init
- identity
- sign
- package
- notarize
- verify
- publish
- release

Internal Systems

- apple
- keychain
- signing
- package
- notarize
- github
- ci
- config
- verification

## Security

Private keys never leave local ownership except encrypted CI export.

Dedicated keychains only.

No login.keychain usage.

Full audit trail.

## Long Term Vision

Become the open standard for macOS release trust infrastructure.

Remove operational friction.

Remain permanently free.

No paywalls.

Civilization-grade software.
