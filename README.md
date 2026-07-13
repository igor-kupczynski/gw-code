# gw-code

Open any Grove multi-repository workspace in VS Code (or Cursor with a one-line config override).

## Install

Install [Grove](https://github.com/nicksenap/grove) and VS Code's `code` CLI on your `PATH`:

```bash
brew install nicksenap/grove/grove
```

Released plugin:

```bash
gw plugin install igor-kupczynski/gw-code
```

### Local checkout

```bash
mkdir -p bin ~/.grove/plugins
go build -o bin/gw-code ./cmd/gw-code
ln -sf "$PWD/bin/gw-code" ~/.grove/plugins/gw-code
gw code --version
```

After code changes:

```bash
go build -o bin/gw-code ./cmd/gw-code
```

Remove the symlink with `rm ~/.grove/plugins/gw-code`.

## Quick start

```bash
# From inside a Grove workspace: generate and open
gw code

# From anywhere
gw code my-workspace

# Generate only, print path
gw code my-workspace --refresh

# Print deterministic path without generating or opening
gw code my-workspace --path
```

`--refresh` and `--path` are mutually exclusive. Every launch opens a **new window** (`--new-window`). Shell completion is not currently provided; Grove invokes the plugin as `gw code`, so standalone completion would not match real usage.

## First-time Grove example

Use [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Bubbles](https://github.com/charmbracelet/bubbles):

```bash
mkdir -p ~/Code/grove-demo
git clone --depth 1 https://github.com/charmbracelet/bubbletea.git ~/Code/grove-demo/bubbletea
git clone --depth 1 https://github.com/charmbracelet/bubbles.git ~/Code/grove-demo/bubbles
gw init ~/Code/grove-demo
gw repos
gw create editor-demo --repos bubbletea,bubbles --branch try-gw-code
gw code editor-demo
```

`gw code` generates `<workspace>/<workspace-name>.code-workspace` with one folder per Grove repo and opens it with VS Code.

## Editor configuration

Default executable is always `code`. Configure a different executable and its arguments globally in `~/.grove/gw-code.toml` (or `$GROVE_DIR/gw-code.toml`):

```toml
# ~/.grove/gw-code.toml
editor = "cursor"
args = ["--classic"]
```

`gw-code` passes configured arguments before the required `--new-window` flag and workspace path:

```text
cursor --classic --new-window <workspace>/editor-demo.code-workspace
```

Shell aliases such as `cr` are not visible to direct process execution; configure the actual executable name and arguments instead. There is no per-workspace config or CLI override.

If the editor binary is missing, `gw-code` still generates the workspace file and prints its path on stderr before returning the launch error.

## Generated files

```text
<workspace>/
  <workspace-name>.code-workspace   # generated; may be overwritten
```

`gw-code` never modifies Grove `state.json` or source repositories.

## Contributing

Best-effort project. Focused changes welcome; no support or process promises.

```bash
go test -race ./...
GW_CODE_INTEGRATION=1 go test -race ./internal/integration -count=1
```

## License

MIT. See [LICENSE](LICENSE).
