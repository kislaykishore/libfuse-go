# Go FUSE Filesystem

This is a simple FUSE filesystem implemented in Go using the low-level `libfuse` API. It creates a filesystem with a single read-only file named `hello` that contains the text "Hello, World!".

## Prerequisites

- Go 1.18 or later
- `libfuse-dev` package (or equivalent for your distribution)

## Building

To build the filesystem, run the following command from the `fuse-go` directory:

```bash
go build -o fuse-go main.go
```

## Running

1. Create a mount point directory:

   ```bash
   mkdir /tmp/fuse-mount
   ```

2. Run the filesystem, passing the mount point as an argument:

   ```bash
   ./fuse-go /tmp/fuse-mount
   ```

3. In a separate terminal, you can now interact with the mounted filesystem:

   ```bash
   ls -l /tmp/fuse-mount
   cat /tmp/fuse-mount/hello
   ```

4. To unmount the filesystem, press Ctrl+C in the terminal where the `fuse-go` program is running.
