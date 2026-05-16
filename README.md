# Git-tool

Windows utility to automate Git + SSH setup for GitHub.

[Releases](https://github.com/nameIess/Git-tool/releases) • [Issues](https://github.com/nameIess/Git-tool/issues)

## ✨ Highlights

- Interactive setup wizard (TUI)
- SSH key generation & agent setup
- Auto-apply Git global config (name/email)
- Upload SSH public key to GitHub

## ⚡ Quick Start

1. Download: [git-tool.exe](https://github.com/nameIess/Git-tool/releases/download/v2.0.0/git-tool.exe)
2. Run:

```powershell
.\git-tool.exe
```

Follow the prompts to complete setup (Git config, SSH keys, connection test).

## 🛠 Build

```powershell
git clone https://github.com/nameIess/Git-tool.git
cd Git-tool
go build -o git-tool.exe .
```

Or run build.bat on Windows.

## 🐞 Quick Troubleshooting

- Ensure `git` is in PATH.
- Back up `~/.ssh` before running.
- For GitHub API: token needs `admin:public_key` scope.
