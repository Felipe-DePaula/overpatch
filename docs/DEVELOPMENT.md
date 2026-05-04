# Development

## Prerequisites

- Go installed and available in `PATH`

Check:

```powershell
go version
where go
```

On Windows, Go can be installed with:

```powershell
winget install GoLang.Go
```

After installing Go, close and reopen your terminal before running build commands.

Common Windows install path:

```text
C:\Program Files\Go\bin
```

## Build check

From the repository root:

```powershell
go test ./...
go build -a -o .\overpatch.exe .\cmd\overpatch
.\overpatch.exe version
Remove-Item .\overpatch.exe -ErrorAction SilentlyContinue
```

## Windows executable manifest

On Windows, executables with names containing words such as `patch`, `update`, `setup`, or `install` may trigger User Account Control heuristics if they do not include an application manifest.

Because the CLI binary is named `overpatch.exe`, the Windows build embeds a manifest declaring that the executable should run with the current user's privileges:

```xml
<requestedExecutionLevel level="asInvoker" uiAccess="false" />
```

This prevents Windows from treating Overpatch like an installer or system patcher that may require elevation.

## Manifest files

The manifest source lives at:

```text
cmd/overpatch/overpatch.exe.manifest
```

The compiled Windows resource file lives at:

```text
cmd/overpatch/rsrc_windows_amd64.syso
```

The `.syso` file is intentionally committed because Go automatically links it into the Windows executable during build.

## Regenerating the Windows resource

Install the resource compiler:

```powershell
go install github.com/akavel/rsrc@latest
```

Regenerate the `.syso` file:

```powershell
rsrc -manifest .\cmd\overpatch\overpatch.exe.manifest -o .\cmd\overpatch\rsrc_windows_amd64.syso
```

If `rsrc` is not available in the current shell path, run it directly from the Go bin directory:

```powershell
& "$env:USERPROFILE\go\bin\rsrc.exe" -manifest .\cmd\overpatch\overpatch.exe.manifest -o .\cmd\overpatch\rsrc_windows_amd64.syso
```

## Building locally on Windows

From the repository root:

```powershell
go clean -cache
go build -a -o .\overpatch.exe .\cmd\overpatch
.\overpatch.exe version
```

Expected output:

```text
overpatch dev
```

## Git ignore guidance

Local build outputs should not be committed:

```gitignore
# Local builds
/overpatch.exe
/*.exe
```

Do not ignore or delete this file unless the build process changes:

```text
cmd/overpatch/rsrc_windows_amd64.syso
```

## Notes

This Windows-specific manifest is only a build resource. It does not change Overpatch behavior, permissions, or security model. Overpatch should run as the current user and should not require administrator privileges for normal CLI usage.

## Environment boundary for agents

Agents and coding tools must operate only inside the repository root.

If Go is not available in `PATH`, agents must report that fact and stop. They must not search user directories or system profile directories to locate Go.

Do not inspect, list, read, or modify paths outside the repository root, including:

- user home directories
- AppData
- WindowsApps
- Downloads
- Documents
- Desktop
- OneDrive
- parent directories
- sibling directories

Use only repository-relative paths unless the user explicitly authorizes otherwise.
