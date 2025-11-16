# Secrets Management

Managing sensitive information in dotfiles requires careful consideration. This guide explains approaches for handling secrets, credentials, and private keys while using dot.

## The Challenge

Dotfiles often contain sensitive information:

- API tokens and credentials
- SSH private keys
- GPG keyrings
- Environment variables with secrets
- Service-specific credentials

Committing these to version control creates security risks. This guide covers best practices for managing secrets alongside your dotfiles.

## Why dot Does Not Implement Encryption

dot intentionally does not implement encryption or secrets management for several reasons:

1. **The Bootstrap Paradox**: GPG keys and SSH keys are themselves dotfiles that need management. Encrypting the encryption keys creates a circular dependency.

2. **Security Responsibility**: Bugs in encryption code could be catastrophic. Corrupting a user's GPG keyring or SSH keys could lock them out of critical systems.

3. **Existing Tools**: Battle-tested tools like `pass`, `gpg`, and `age` already solve this problem well.

4. **Unix Philosophy**: Each tool should do one thing well. dot manages symlinks, existing tools manage secrets.

## What dot Does Automatically

dot protects against accidental secrets management through:

### Default Ignore Patterns

The following are automatically ignored when managing packages:

- `.gnupg` - GPG keyring directory
- `.ssh/id_*` - SSH identity files (private and public keys)
- `.ssh/*.pem` - PEM format keys
- `.ssh/*_rsa` - RSA keys
- `.ssh/*_ecdsa` - ECDSA keys
- `.ssh/*_ed25519` - Ed25519 keys
- `.password-store` - pass password store

These patterns prevent accidentally managing sensitive files. SSH configuration files like `.ssh/config`, `.ssh/known_hosts`, and `.ssh/authorized_keys` are not ignored and can be safely managed.

### Secrets Detection

dot warns when you attempt to manage or adopt files that match sensitive patterns:

```bash
$ dot manage ssh

Warning: Potential secrets detected:
  - /home/user/.ssh/id_rsa (SSH private and public keys)
  - /home/user/.ssh/id_ed25519 (SSH private and public keys)

These files are ignored by default. See 'dot help secrets' for details.
```

The warning is informational only. The files are automatically ignored during management.

## Recommended Approaches

### Approach 1: Using pass (Password Store)

`pass` is a Unix password manager that integrates naturally with dotfiles and git.

#### Setup

```bash
# Initialize pass with your GPG key
pass init your-gpg-key-id

# Enable git integration
pass git init
```

#### Storing Secrets

```bash
# Store a GitHub token
pass insert api/github-token

# Store AWS credentials
pass insert aws/access-key
pass insert aws/secret-key

# Store environment variables
pass insert env/production
```

#### Using in Dotfiles

Reference secrets from your dotfiles:

```bash
# In .bashrc or .zshrc
export GITHUB_TOKEN=$(pass api/github-token)
export AWS_ACCESS_KEY_ID=$(pass aws/access-key)
export AWS_SECRET_ACCESS_KEY=$(pass aws/secret-key)
```

#### Version Control

```bash
# Navigate to password store
cd ~/.password-store

# Push to private repository
git remote add origin git@github.com:user/password-store.git
git push -u origin main
```

Your passwords are GPG-encrypted before committing, so the repository is safe even if exposed.

### Approach 2: Selective GPG Encryption

For specific files containing secrets, use GPG encryption directly.

#### Encrypting Files

```bash
# Encrypt a file
gpg --encrypt --recipient your-email@example.com secrets.env

# Commit encrypted version
git add secrets.env.gpg
git commit -m "Add encrypted secrets"
```

#### Add to .dotignore

Create a `.dotignore` file in your package:

```
# .dotfiles/myapp/.dotignore
secrets.env       # Ignore decrypted version
```

#### Shell Initialization

Decrypt during shell initialization:

```bash
# In .bashrc or .zshrc
if [ -f ~/.dotfiles/myapp/secrets.env.gpg ]; then
    # Decrypt to temporary location
    gpg --quiet --decrypt ~/.dotfiles/myapp/secrets.env.gpg > /tmp/secrets.env
    source /tmp/secrets.env
    rm /tmp/secrets.env
fi
```

### Approach 3: Using age for Encryption

`age` is a modern, simple encryption tool that works well with dotfiles.

#### Setup

```bash
# Install age
# macOS: brew install age
# Linux: apt install age

# Generate a key
age-keygen -o ~/.age-key.txt

# Display public key
cat ~/.age-key.txt
```

#### Encrypting Files

```bash
# Encrypt with your public key
age --encrypt --recipient age1... secrets.env > secrets.env.age

# Decrypt
age --decrypt --identity ~/.age-key.txt secrets.env.age > secrets.env
```

#### Integration

```bash
# In .bashrc or .zshrc
if [ -f ~/.dotfiles/myapp/secrets.env.age ]; then
    age --decrypt --identity ~/.age-key.txt \
        ~/.dotfiles/myapp/secrets.env.age > /tmp/secrets.env
    source /tmp/secrets.env
    rm /tmp/secrets.env
fi
```

### Approach 4: Environment-Specific Files

Keep sensitive configuration out of version control entirely.

#### Structure

```
.dotfiles/
├── myapp/
│   ├── config.yaml          # Safe to commit
│   └── secrets.yaml         # Never commit (in .gitignore)
└── .gitignore
```

#### Ignore Secrets

```
# .dotfiles/.gitignore
secrets.yaml
secrets.env
*.secret
```

#### Template Approach

Commit example files:

```
.dotfiles/
├── myapp/
│   ├── config.yaml
│   ├── secrets.yaml.example    # Template
│   └── secrets.yaml            # Actual (ignored)
```

Users copy the example and fill in their own values:

```bash
cp secrets.yaml.example secrets.yaml
# Edit secrets.yaml with actual values
```

## The Bootstrap Problem

### GPG Keys

The `.gnupg` directory contains your private keys and should be:

1. **Set up manually** on new machines
2. **Never committed** to version control
3. **Backed up separately** using GPG-specific tools:

```bash
# Export private key (encrypted)
gpg --export-secret-keys --armor your-email > private-key.asc

# Import on new machine
gpg --import private-key.asc
```

### SSH Keys

Similarly, `.ssh` keys should be:

1. **Generated per-machine** or securely transferred
2. **Never committed** to version control
3. **Protected with strong passphrases**

```bash
# Generate new key
ssh-keygen -t ed25519 -C "your-email@example.com"

# Or securely copy from another machine
scp old-machine:.ssh/id_ed25519 ~/.ssh/
chmod 600 ~/.ssh/id_ed25519
```

The SSH configuration file (`.ssh/config`) is safe to manage with dot.

## Best Practices

### Do

- Use dedicated secrets management tools (pass, gpg, age)
- Keep GPG and SSH keys outside version control
- Encrypt before committing sensitive files
- Use `.dotignore` to exclude secrets from packages
- Generate per-machine keys when possible
- Use strong passphrases for encryption keys
- Test decryption scripts to avoid lockout
- Document secret locations in README

### Do Not

- Commit unencrypted secrets to version control
- Manage `.gnupg` or `.ssh/id_*` with dot
- Disable default ignore patterns for security files
- Share private keys between machines unnecessarily
- Store encryption keys in the same repository
- Skip testing restore procedures

## Configuration Integration

Configure per-package ignore patterns if needed:

```yaml
# ~/.config/dot/config.yaml
ignore:
  use_defaults: true
  patterns:
    - "*.secret"
    - "credentials.*"
    - ".env.local"
```

## Troubleshooting

### Secrets Were Committed

If secrets were accidentally committed:

```bash
# Remove from history (requires force push)
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch path/to/secret' \
  --prune-empty --tag-name-filter cat -- --all

# Rotate the compromised secrets immediately
```

### Can't Access Encrypted Files

If you lose access to encrypted files:

1. Check GPG/age key backups
2. Verify key expiration dates
3. Check file permissions on key files
4. Test decryption with known-good files

### Ignore Patterns Not Working

Verify patterns are correctly formatted:

```bash
# Test ignore patterns
dot manage mypackage --dry-run

# Check what files would be managed
```

## See Also

- [Ignore System](ignore-system.md) - Detailed ignore pattern documentation
- [Configuration Guide](04-configuration.md) - Configuration options
- [Commands Reference](05-commands.md) - Command usage

## External Resources

- [pass](https://www.passwordstore.org/) - The standard Unix password manager
- [age](https://age-encryption.org/) - Simple, modern file encryption
- [GPG](https://gnupg.org/) - GNU Privacy Guard
- [git-secret](https://git-secret.io/) - Store secrets in git repositories
- [SOPS](https://github.com/mozilla/sops) - Secrets OPerationS (Mozilla)

