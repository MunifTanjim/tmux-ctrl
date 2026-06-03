# tmux-ctrl

A command-line tool for controlling tmux.

## Installation

**Prerequisites**:

- [GitHub CLI](https://cli.github.com/) (`gh`)

```sh
brew install gh

gh auth login
```

**Quick Install**:

```sh
gh api -H "Accept: application/vnd.github.raw" repos/MunifTanjim/tmux-ctrl/contents/scripts/install.sh | bash
```

This will install to `~/.local/bin/tmux-ctrl`.

To install to a custom directory:

```sh
gh api -H "Accept: application/vnd.github.raw" repos/MunifTanjim/tmux-ctrl/contents/scripts/install.sh | INSTALL_DIR=$HOME/.local/bin bash
```

**Manual Installation (from source)**:

Requires Go 1.26+.

1. Clone the repository:

   ```sh
   git clone https://github.com/MunifTanjim/tmux-ctrl.git
   cd tmux-ctrl
   ```

2. Install the CLI:

   ```sh
   make install
   ```

   This will build the binary and install it to `~/.local/bin/tmux-ctrl`.

### Shell Completion

```sh
tmux-ctrl completion install
# or
tmux-ctrl completion zsh > "${fpath[1]}/_tmux-ctrl"
```

## License

Licensed under the MIT License. Check the [LICENSE](./LICENSE) file for details.
