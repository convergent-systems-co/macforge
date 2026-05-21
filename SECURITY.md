# Security Policy

MacForge handles signing identities and (indirectly) private keys. We
take security seriously.

## Reporting a vulnerability

Email **security@convergent.systems** with:

- A description of the issue.
- A minimal reproduction (please don't include real signing material).
- Your contact info for follow-up.

We aim to acknowledge within 72 hours and provide a fix or mitigation
within 30 days for confirmed vulnerabilities.

## Supported versions

While in v0.x, only the latest minor release receives security fixes.

## Threat model — what's in scope

- Argv leakage of private keys / keychain passwords into logs or audit.
- Bypasses of the keychain-name validator that touch `login.keychain`.
- Bypasses of the inline-password-rejection validator.
- Issues that cause MacForge to sign artifacts under the wrong identity.
- Audit-log integrity (e.g., events lost or reordered).

## What's out of scope

- Vulnerabilities in Apple's CLI tools themselves (report to Apple).
- Vulnerabilities in cobra/viper/zerolog (report upstream).
- General macOS Gatekeeper / notarization weaknesses (report to Apple).

## Disclosure

We follow coordinated disclosure. We will credit the reporter in the
release notes unless they request anonymity.
