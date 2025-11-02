# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in `dot`, please report it by emailing the maintainers or opening a private security advisory on GitHub. Please do not open public issues for security vulnerabilities.

## Known Vulnerabilities

### GO-2024-3295: GitHub CLI Authentication in Codespaces

**Status**: Tracked, waiting for upstream fix  
**Severity**: Low (for this project)  
**Affected Package**: `github.com/cli/go-gh@v1.2.1`  
**Description**: Violation of GitHub host security boundary when sourcing authentication token within a codespace.

**Impact Assessment**:
- ✅ **Low Risk**: This vulnerability specifically affects GitHub Codespaces environments
- ✅ **Limited Scope**: `dot` is a CLI tool primarily used in local development environments
- ✅ **Fallback Available**: The tool has multiple authentication fallback methods (environment variables, SSH keys, no auth)
- ✅ **Optional Feature**: GitHub CLI authentication is only one of several authentication options

**Usage in `dot`**:
The affected code is in `internal/adapters/git_auth.go:92` where `auth.TokenForHost("github.com")` is called as a convenience feature to automatically detect GitHub CLI credentials for HTTPS clones. This is:
1. **Priority 3** in the authentication chain (after env vars and SSH keys)
2. **Optional** - users can use `GITHUB_TOKEN` env var or SSH keys instead
3. **Gracefully degrades** - falls back to public/no-auth if unavailable

**Mitigation**:
Users concerned about this vulnerability in codespace environments can:
1. Use `GITHUB_TOKEN` environment variable (priority 1)
2. Use SSH URLs with SSH key authentication (priority 2)
3. Avoid using GitHub CLI authentication in codespaces

**Resolution Plan**:
- Monitor `github.com/cli/go-gh` for security updates
- Update dependency immediately when fix is available
- Current version v1.2.1 is the latest available (as of dependency check)

**References**:
- https://pkg.go.dev/vuln/GO-2024-3295
- Issue tracking: [To be created if needed]

## Security Best Practices

When using `dot`:

1. **Authentication Priority**:
   - Use `GITHUB_TOKEN` environment variable for automated environments
   - Use SSH keys for local development
   - GitHub CLI detection is a convenience feature, not required

2. **Permissions**:
   - Review the permissions required for your dotfiles repository
   - Use read-only tokens when possible
   - Consider using deploy keys for specific repositories

3. **Sensitive Data**:
   - Never commit secrets, tokens, or credentials to your dotfiles
   - Use `.gitignore` patterns to exclude sensitive files
   - Review your dotfiles before making repositories public

4. **Updates**:
   - Keep `dot` updated to receive security patches
   - Run `dot upgrade` regularly
   - Monitor the changelog for security-related updates

## Vulnerability Disclosure Timeline

When security vulnerabilities are discovered and fixed:
1. Patch is developed and tested
2. Security advisory is published
3. Fixed version is released
4. Public announcement after users have time to upgrade (typically 7-14 days)


