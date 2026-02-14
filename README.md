# Git Tool 🚀

A user-friendly Windows terminal application that automates the setup of Git and SSH for GitHub. This tool guides you through the entire process of configuring Git, generating SSH keys, adding them to GitHub, and testing the connection.

[📦 Releases](https://github.com/nameIess/Git-tool/releases) • [🐛 Issues](https://github.com/nameIess/Git-tool/issues)

---

## ✨ Features

- 🔍 **Prerequisites Check**: Verifies that Git and SSH are installed and available
- 🧑‍💻 **Git Configuration**: Interactive setup of Git user name and email
- 🔐 **SSH Key Generation**: Automatically generates a new SSH key pair
- 🤖 **SSH Agent Setup**: Configures SSH agent to manage your keys
- 🌐 **GitHub Integration**: Adds your SSH key to GitHub via the API
- 🧪 **Connection Testing**: Validates the SSH connection to GitHub
- 📊 **Progress Tracking**: Step-by-step progress with clear status indicators
- ⚠️ **Error Handling**: Comprehensive error reporting and logging
- 🪟 **Windows Optimized**: Built specifically for Windows with proper shell detection

---

## ⚡ Quick Start

1. **Download** the latest release from [Releases Page](https://github.com/nameIess/Git-tool/releases) or build from source:

   ```bash
   go build -o git-tool.exe .
   ```

2. **Run** the executable:

   ```bash
   .\git-tool.exe
   ```

3. **Follow** the on-screen prompts through each phase:
   - Prerequisites Check
   - Git Configuration
   - SSH Key Generation
   - SSH Agent Setup
   - GitHub Integration
   - Connection Test

That's it! 🎉 Your Git and SSH setup is complete.

---

## 📦 Installation

### Prerequisites

- 🪟 Windows 10 or later
- 🛠️ Git installed and available in PATH
- 🔑 SSH client (usually included with Git for Windows)
- 🌐 Internet connection for GitHub API access

### Download

Download the latest release from the [Releases page](https://github.com/nameIess/Git-tool/releases).

**Direct Downloads:**
- [📥 git-tool.exe (v1.0)](https://github.com/nameIess/Git-tool/releases/download/v1.0/git-tool.exe) - Windows executable
- [📦 Source Code (ZIP)](https://github.com/nameIess/Git-tool/archive/refs/heads/main.zip) - Source code archive

### Build from Source

1. Install Go 1.25.6 or later
2. Clone this repository:
   ```bash
   git clone https://github.com/nameIess/Git-tool.git
   cd Git-tool
   ```
3. Build the application:
   ```bash
   go build -o git-tool.exe .
   ```
   Or use the provided build script:
   ```bash
   .\build.bat
   ```

---

## ▶️ Usage

1. Run the executable:

   ```bash
   .\git-tool.exe
   ```

2. Follow the on-screen prompts through each phase:
   - **Prerequisites Check**: Verifies Git and SSH installation
   - **Git Configuration**: Set your name and email
   - **SSH Key Generation**: Creates a new SSH key pair
   - **SSH Agent Setup**: Configures the SSH agent
   - **GitHub Integration**: Adds your key to GitHub
   - **Connection Test**: Validates the setup

3. The tool will guide you through each step with clear instructions and status indicators.

### Phases

#### 1. 🔍 Prerequisites Check

- Verifies Git installation and version
- Checks SSH availability
- Detects shell environment (cmd, PowerShell, Git Bash)

#### 2. 🧑‍💻 Git Configuration

- Prompts for Git user name
- Prompts for Git user email
- Validates input format
- Applies configuration to global Git settings

#### 3. 🔐 SSH Key Generation

- Generates a new Ed25519 SSH key pair
- Saves keys to `~/.ssh/id_ed25519`
- Displays the public key for reference

#### 4. 🤖 SSH Agent Setup

- Starts or connects to the SSH agent
- Adds the new SSH key to the agent
- Configures automatic key loading

#### 5. 🌐 GitHub Integration

- Prompts for GitHub username
- Prompts for GitHub personal access token
- Adds the SSH key to your GitHub account via API
- Validates the token and username

#### 6. 🧪 Connection Test

- Tests SSH connection to GitHub
- Displays your GitHub username if successful
- Provides troubleshooting information if failed

---

## 🔧 Configuration

### Git Configuration

The tool sets the following Git configuration:

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### SSH Configuration

Creates or updates `~/.ssh/config` with:

```
Host github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/id_ed25519
```

---

## 📝 Logging

The application creates a detailed log file in the same directory as the executable:

- `git-setup-YYYY-MM-DD_HH-MM-SS.log` - Contains all operations, errors, and system information
- Useful for troubleshooting and support

---

## 🐞 Troubleshooting

### Common Issues

**Git not found in PATH** 🔍

- Ensure Git for Windows is installed
- Add Git to your system PATH environment variable

**SSH agent not running** 🤖

- The tool will attempt to start the SSH agent automatically
- If issues persist, try running `ssh-agent` manually

**GitHub API rate limiting** 🌐

- Ensure you're using a valid personal access token
- Check that the token has the `admin:public_key` scope

**Permission denied errors** 🚫

- Verify your SSH key was added to GitHub successfully
- Check that the key file permissions are correct (600 for private key)

### Manual SSH Key Setup

If the automated setup fails, you can manually create an SSH key:

```bash
ssh-keygen -t ed25519 -C "your.email@example.com"
```

Then add the public key (`~/.ssh/id_ed25519.pub`) to your GitHub account manually.

---

## 🔒 Security

- 🔐 SSH keys are generated locally and never transmitted
- 🛡️ GitHub personal access tokens are only used for API authentication
- 📝 All operations are logged for transparency
- ✅ The tool follows GitHub's security best practices

---

## 👩‍💻 Development

### Requirements

- Go 1.25.6 or later
- Git
- SSH client

### Building

```bash
go mod tidy
go build -o git-tool.exe .
```

### Testing

The application includes comprehensive error handling and validation. To test:

1. Run the application in a clean environment
2. Test each phase individually
3. Verify error handling for missing prerequisites
4. Test GitHub API integration with valid and invalid tokens

### Architecture

The application uses the Bubble Tea framework for the TUI and is organized into phases:

- `internal/tui/` - Terminal user interface components
- `internal/shell/` - Shell and system interaction
- `internal/logger/` - Logging functionality
- `internal/exec/` - Command execution utilities

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
