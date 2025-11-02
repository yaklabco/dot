# Security Vulnerability Exceptions

This document tracks known security vulnerabilities that are intentionally excluded from CI checks, along with the rationale for each exception.

## Active Exceptions

### GO-2024-3295: GitHub CLI Host Security Boundary Violation

**Vulnerability ID:** GO-2024-3295  
**Affected Package:** `github.com/cli/go-gh@v1.2.1`  
**Severity:** Medium  
**Status:** Accepted (not fixing)

#### Description

Violation of GitHub host security boundary when sourcing authentication token within a GitHub Codespace.

**References:**
- https://pkg.go.dev/vuln/GO-2024-3295
- https://github.com/cli/go-gh

#### Impact Assessment

- **Scope:** Only affects code running inside GitHub Codespaces
- **Our Usage:** We use `github.com/cli/go-gh` in `internal/adapters/git_auth.go` to retrieve GitHub CLI tokens for repository authentication
- **Risk Level:** **Low** for typical usage
  - Most users run `dot` locally or in CI, not in Codespaces
  - The vulnerability requires specific Codespace conditions
  - Only affects authentication token retrieval, not core functionality
  - Users can still use environment variables (`GITHUB_TOKEN`, `GIT_TOKEN`) or SSH keys as alternatives

#### Rationale for Exception

1. **Limited Scope**: Vulnerability only affects GitHub Codespaces environment
2. **Workarounds Available**: Multiple authentication methods supported:
   - `GITHUB_TOKEN` environment variable (priority 1)
   - `GIT_TOKEN` environment variable (priority 2)
   - SSH keys (priority 3)
   - GitHub CLI tokens (priority 4) ← affected
   - Public repository access (no auth)
3. **No Fix Available**: Package maintainers have not released a patched version
4. **Low User Impact**: Most users don't use Codespaces with our tool
5. **Code Path**: Only triggered when:
   - Running in a Codespace
   - No environment token variables set
   - Using HTTPS GitHub URLs (not SSH)
   - GitHub CLI is authenticated

#### Mitigation

For users running in GitHub Codespaces:
1. Set `GITHUB_TOKEN` environment variable explicitly
2. Use SSH URLs instead of HTTPS URLs
3. Use SSH key authentication

#### Monitoring

- Reviewed: 2025-11-02
- Next Review: When `github.com/cli/go-gh` releases a fix
- Action: Remove exception when patched version available

#### Configuration

Exception configured in:
- `Makefile` - `vuln` target filters GO-2024-3295
- `.github/workflows/ci.yml` - Vulnerability check excludes GO-2024-3295
- `.github/workflows/version-bump.yml` - Uses Makefile vuln target with exclusion

---

## Exception Process

When adding a new exception:

1. **Document the vulnerability**:
   - ID, package, severity
   - Detailed description
   - Impact on our codebase

2. **Justify the exception**:
   - Why not fixing immediately?
   - What's the risk assessment?
   - What workarounds exist?

3. **Update configurations**:
   - Makefile vuln target
   - CI workflows
   - This document

4. **Set review schedule**:
   - When to re-evaluate?
   - What triggers removal?

5. **Get approval**:
   - Security review
   - Maintainer sign-off

## Review Schedule

Security exceptions should be reviewed:
- **Monthly**: Check for updated package versions with fixes
- **Pre-release**: Verify exceptions are still valid
- **Quarterly**: Full security audit of all exceptions

Last full review: 2025-11-02
Next review: 2025-12-02

