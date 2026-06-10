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

## Usage

### Extract (hint overlay)

`tmux-ctrl pane extract` finds tokens (URLs, paths, git SHAs, and any
config-defined patterns) in a pane and prints the selection. With `--hint` it
shows a tmux-thumbs-style overlay: the pane content is dimmed and a hint key is
drawn at the start of each match; type the hint to select it.

Bind it to a tmux key so the overlay pops up over the current pane and copies the
pick to the tmux buffer:

```tmux
bind-key F display-popup -E -B \
  -x '#{pane_left}' -y '#{pane_top}' -w '#{pane_width}' -h '#{pane_height}' \
  "tmux-ctrl pane extract --hint --copy -p '#{pane_id}'"
```

Run directly in a pane to print instead (composable):

```sh
tmux-ctrl pane extract --hint | pbcopy
```

## License

Licensed under the MIT License. Check the [LICENSE](./LICENSE) file for details.
