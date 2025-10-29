# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of ckeletin-go seriously. If you have discovered a security vulnerability, we appreciate your help in disclosing it to us in a responsible manner.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **Email**: Send an email to the project maintainer
2. **GitHub Security Advisories**: Use the [GitHub Security Advisory](https://github.com/peiman/ckeletin-go/security/advisories/new) feature

### What to Include

Please include the following information in your report:

- Type of vulnerability (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: Within 48 hours, you will receive an acknowledgment of your report
- **Status Update**: Within 7 days, we will send a detailed response indicating the next steps
- **Fix Timeline**: We aim to release security patches within 30 days of confirmation
- **Disclosure**: We follow coordinated vulnerability disclosure practices

### What to Expect

After submitting a report, you can expect:

1. **Confirmation** - We will confirm receipt of your vulnerability report
2. **Investigation** - We will investigate and validate the reported vulnerability
3. **Resolution** - We will work on a fix and determine the release timeline
4. **Credit** - We will credit you in the security advisory (unless you prefer to remain anonymous)
5. **Notification** - We will notify you when the vulnerability is fixed and publicly disclosed

## Security Best Practices for Users

When using ckeletin-go in your projects:

1. **Keep Updated**: Always use the latest stable release
2. **Monitor Advisories**: Watch this repository for security advisories
3. **Dependency Scanning**: Use tools like `govulncheck` to scan your dependencies
4. **Input Validation**: Always validate user input in your CLI applications
5. **Least Privilege**: Run CLI applications with minimal necessary permissions

## Automated Security Measures

This project uses several automated security measures:

- **Dependabot**: Automatic dependency updates for known vulnerabilities
- **CodeQL**: Automated code scanning for security issues
- **govulncheck**: Regular vulnerability scanning of Go dependencies
- **SBOM Generation**: Software Bill of Materials for transparency

## Security-Related Configuration

### Safe Configuration Practices

When configuring ckeletin-go-based applications:

- Use environment variables for sensitive data (never commit secrets)
- Validate all configuration inputs
- Use secure defaults
- Follow the principle of least privilege

## Attribution

We would like to publicly thank the following people for responsibly disclosing security vulnerabilities:

<!-- Security researchers will be credited here -->

*None yet - be the first to help improve security!*

## Policy Updates

This security policy may be updated from time to time. Please check back regularly for any changes.

**Last Updated**: 2025-10-29
