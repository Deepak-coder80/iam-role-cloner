# IAM Role Cloner

🚀 A powerful CLI tool to clone IAM roles between AWS environments with pattern replacement and comprehensive logging.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Deepak-coder80/iam-role-cloner.svg)](https://github.com/Deepak-coder80/iam-role-cloner/releases)

## ✨ Features

- 🔄 **Clone IAM roles** between different AWS accounts/profiles
- 🔧 **Pattern replacement** in role names, policy content, and tags
- 🧪 **Dry-run mode** for safe testing before making changes
- 📝 **Comprehensive logging** with colored output and file logging
- 🔍 **Role discovery** with automatic pattern-based filtering
- 🛡️ **Error handling** with detailed error messages and rollback guidance
- 📊 **Progress tracking** for batch operations
- 🎯 **Interactive mode** or full command-line automation
- 🏷️ **Tag management** with automatic environment tag updates

## 🚀 Quick Start

### Installation

#### Download Pre-built Binaries

```bash
# Linux (AMD64)
curl -L -o iam-role-cloner https://github.com/Deepak-coder80/iam-role-cloner/releases/latest/download/iam-role-cloner-linux-amd64
chmod +x iam-role-cloner

# macOS (AMD64)
curl -L -o iam-role-cloner https://github.com/Deepak-coder80/iam-role-cloner/releases/latest/download/iam-role-cloner-darwin-amd64
chmod +x iam-role-cloner

# Windows (AMD64)
curl -L -o iam-role-cloner.exe https://github.com/Deepak-coder80/iam-role-cloner/releases/latest/download/iam-role-cloner-windows-amd64.exe
```

#### Build from Source

```bash
git clone https://github.com/Deepak-coder80/iam-role-cloner.git
cd iam-role-cloner
go build -o iam-role-cloner .
```

### Prerequisites

- AWS CLI configured with appropriate profiles
- IAM permissions for source and destination accounts:
  - `iam:GetRole`
  - `iam:CreateRole`
  - `iam:ListRoles`
  - `iam:ListAttachedRolePolicies`
  - `iam:ListRolePolicies`
  - `iam:GetRolePolicy`
  - `iam:AttachRolePolicy`
  - `iam:PutRolePolicy`
  - `iam:TagRole`
  - `iam:ListRoleTags`

## 📚 Usage

### Interactive Mode (Recommended for beginners)

```bash
# Launch interactive wizard
./iam-role-cloner clone

# Dry-run with verbose output
./iam-role-cloner clone --dry-run --verbose
```

### Command Line Mode (Automation-friendly)

```bash
# Clone with all parameters specified
./iam-role-cloner clone \
  --source-profile dev \
  --dest-profile prod \
  --source-pattern "dev_" \
  --dest-pattern "prod_" \
  --dry-run

# List roles in a profile
./iam-role-cloner list --profile dev --pattern "app"

# Show version information
./iam-role-cloner version --detailed
```

## 🛠️ Commands

### `clone` - Clone IAM Roles

Clone IAM roles from source to destination with pattern replacement.

```bash
./iam-role-cloner clone [flags]
```

**Flags:**
- `-s, --source-profile` - Source AWS profile
- `-d, --dest-profile` - Destination AWS profile
- `--source-pattern` - Source environment pattern (e.g., 'dev_')
- `--dest-pattern` - Destination environment pattern (e.g., 'prod_')
- `--dry-run` - Show what would be done without making changes
- `-v, --verbose` - Enable verbose output
- `--log-file` - Custom log file path

**Examples:**

```bash
# Interactive mode
./iam-role-cloner clone

# Fully automated
./iam-role-cloner clone -s dev -d prod --source-pattern "dev_" --dest-pattern "prod_"

# Safe testing
./iam-role-cloner clone --dry-run --verbose
```

### `list` - List IAM Roles

Discover and inspect IAM roles in your AWS accounts.

```bash
./iam-role-cloner list --profile <profile-name> [flags]
```

**Flags:**
- `-p, --profile` - AWS profile (required)
- `--pattern` - Filter roles by pattern
- `--details` - Show detailed role information
- `--sort` - Sort roles alphabetically

**Examples:**

```bash
# List all roles
./iam-role-cloner list --profile dev

# List roles with pattern
./iam-role-cloner list --profile prod --pattern "app" --sort

# Detailed role information
./iam-role-cloner list --profile staging --details
```

### `version` - Version Information

Display version and build information.

```bash
./iam-role-cloner version [--detailed]
```

## 🔧 Configuration Examples

### Basic Environment Migration

Clone roles from development to production:

```bash
./iam-role-cloner clone \
  --source-profile development \
  --dest-profile production \
  --source-pattern "dev-" \
  --dest-pattern "prod-"
```

### Multi-Role Batch Processing

The tool automatically discovers roles and lets you select multiple roles for cloning:

1. **Pattern Discovery**: Finds all roles matching source pattern
2. **Role Selection**: Choose specific roles or select 'all'
3. **Batch Processing**: Clones multiple roles with progress tracking

### Pattern Replacement Examples

| Source Pattern | Dest Pattern | Example Transformation |
|---------------|--------------|----------------------|
| `dev_` | `prod_` | `dev_app_role` → `prod_app_role` |
| `staging-` | `live-` | `staging-api-role` → `live-api-role` |
| `test.` | `prod.` | `test.service.role` → `prod.service.role` |

## 📋 What Gets Cloned

✅ **Trust Policies** (assume role policies) with pattern replacement
✅ **Managed Policies** (AWS and customer managed)
✅ **Inline Policies** with pattern replacement in content
✅ **Tags** with pattern replacement and environment updates
✅ **Role Description** with clone metadata

## 🔒 Security Considerations

- **Dry-run first**: Always test with `--dry-run` before actual cloning
- **Least privilege**: Ensure your AWS credentials have minimal required permissions
- **Cross-account**: Tool works within accounts; cross-account cloning requires appropriate trust relationships
- **Pattern validation**: Review pattern replacements in dry-run output
- **Logging**: All operations are logged for audit trails

## 🐛 Troubleshooting

### Common Issues

**"MalformedPolicyDocument" errors:**
```bash
# Check policies in dry-run mode
./iam-role-cloner clone --dry-run --verbose
```

**"Role already exists" errors:**
- Tool checks for existing roles and prompts for confirmation
- Use different destination patterns to avoid conflicts

**AWS credential errors:**
```bash
# Verify AWS configuration
aws sts get-caller-identity --profile <profile-name>
```

### Debug Mode

Enable verbose logging for detailed troubleshooting:

```bash
./iam-role-cloner clone --verbose --dry-run
```

Log files are automatically created with timestamps for audit purposes.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- Inspired by the need for reliable IAM role management across environments

## 📞 Support

- 📖 [Documentation](https://github.com/Deepak-coder80/iam-role-cloner/wiki)
- 🐛 [Issue Tracker](https://github.com/Deepak-coder80/iam-role-cloner/issues)
- 💬 [Discussions](https://github.com/Deepak-coder80/iam-role-cloner/discussions)

---

**Made with ❤️ for DevOps engineers who value automation and reliability.**
