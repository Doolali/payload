# payload

A desktop HTTP client for exploring APIs — projects, collections, scratchpad sessions, and `{{variable}}` substitution, all stored in plain JSON files on disk.

Built with [Wails](https://wails.io) (Go + Vue 3 + TypeScript).

## Features

- **Projects** group related requests into named **Collections**.
- **Sessions** are an always-available scratchpad list — separate from any project, kept across restarts, with the last response preserved next to each session.
- **Variables** at project and collection scope, referenced anywhere in a request as `{{name}}`. Collection vars shadow project vars of the same name at send time.
- **Extractors** pull values out of a JSON response (`data.token`, `items[0].id`) and write them back as project variables, so a login response can populate the auth token used by every later request.
- **Open** `.json` project files from anywhere on disk, or **move** existing ones into a different folder. Project files are portable — share them like any other source file.
- **JSON viewer** for response bodies, with keyboard-friendly editing for request bodies.
- **UI state persistence** — the last open tab, project, and selection are restored on next launch.

## Install

Download the latest release for your platform from [GitHub Releases](https://github.com/Doolali/payload/releases/latest):

| Platform | Asset |
| --- | --- |
| Windows | `payload-X.Y.Z-windows-amd64.msi` (installs to `C:\Program Files\payload`) |
| Windows (portable) | `payload-X.Y.Z-windows-amd64.zip` |
| macOS | `payload-X.Y.Z-darwin-universal.zip` |
| Linux | `payload-X.Y.Z-linux-amd64.tar.gz` |

### Auto-updates

The app checks GitHub for a newer release shortly after launch and shows a banner if one is available. You can also trigger a check anytime via **Help → Check for Updates...**.

On Windows, accepting the update downloads the new MSI, launches it (you'll see one UAC prompt), and quits the app so the installer can replace the binary cleanly. On macOS and Linux, the updater opens the GitHub release page so you can download the correct asset manually.

## Develop

Requires Go 1.23+ and Node 20+.

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails dev
```

`wails dev` runs the Vite dev server with hot reload for the frontend and rebuilds the Go side on save. Frontend devtools are available at `http://localhost:34115`.

## Build a local binary

```sh
wails build
```

The output lands in `build/bin/`.

## Release

Releases are produced by [`.github/workflows/release.yml`](.github/workflows/release.yml). Push to a `release/MAJOR.MINOR` branch (e.g. `release/1.2`) and the workflow:

1. Picks the next patch version (highest existing `MAJOR.MINOR.*` tag + 1, or `.0`).
2. Builds macOS, Windows, and Linux binaries with the version embedded via `-X main.version=…`.
3. Builds a Windows MSI installer with [WiX](https://wixtoolset.org) — see [`build/windows/installer/payload.wxs`](build/windows/installer/payload.wxs).
4. Tags the commit and publishes a GitHub Release with all artifacts attached.

The auto-updater reads from `/repos/Doolali/payload/releases/latest`, so a published release is immediately discoverable by every installed copy.

## License

GPL v3 — see [LICENSE](LICENSE).
