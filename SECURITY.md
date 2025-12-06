# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by sending an email to:

**seer@mendex.io**

Please include:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Any suggested fixes (optional)

### What to expect

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 7 days
- **Resolution timeline**: Depends on severity

### Please do not

- Open public issues for security vulnerabilities
- Exploit vulnerabilities beyond what's necessary to demonstrate them
- Share vulnerability details before a fix is released

## Security Best Practices

When self-hosting Seer:

1. **Keep updated**: Always run the latest version
2. **Use HTTPS**: Put Seer behind a reverse proxy with TLS
3. **Limit access**: Use firewall rules to restrict access
4. **Secure credentials**: Never commit config.yaml with API keys
5. **Regular backups**: Backup your SQLite database regularly
