# Git Tool GUI rewrite

This branch rewrites Git Tool as a Windows desktop application using Wails v2.

Themes:
- Git Tool Dark
- Windows Native

Both themes share the same Go backend and security workflow. The existing `icon.ico` is the only application icon source.

Development: `wails doctor` then `wails dev`.

Production: `wails build -platform windows/amd64 -webview2 download`.

The setup verifies Git/OpenSSH, configures Git identity, creates or selects an Ed25519 SSH key, configures ssh-agent, verifies GitHub with `ssh -T git@github.com`, and makes SSH signing opt-in.
