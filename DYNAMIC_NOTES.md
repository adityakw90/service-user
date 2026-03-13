# Dynamic Binary Notes

## What is "dynamic"?

The `*_linux_glibc*_*.tar.gz` binaries are dynamically linked against glibc.
This means they require a compatible glibc version on the host system.

## Benefits of Dynamic Binaries

- ✅ Smaller binary size (~15MB vs ~50MB for static)
- ✅ Shared memory usage when running multiple services on same host
- ✅ System-level security updates via glibc patches
- ✅ Better integration with system monitoring tools

## When to Use Dynamic vs Static

### Use Dynamic When:
- Deploying to traditional Linux servers (not containers)
- Running multiple services on the same host
- Memory efficiency is important
- You control the deployment environment

### Use Static When:
- Running in containers (Docker, Kubernetes)
- Deploying to Alpine Linux (uses musl, not glibc)
- Uncertain about target system
- Maximum portability needed

## Available Variants

### glibc228 (Debian 10)
- **Minimum required:** glibc 2.28 or higher
- **Runs on systems with:** glibc ≥2.28 (forward compatible)
- **Compatible distributions:**
  - Debian 10, 11, 12, 13+
  - Ubuntu 20.04, 22.04, 24.04+
  - CentOS 8, Stream 9
  - Fedora 32+
- **Build platform:** Debian 10 (buster)
- **Use when:** You need to support Debian 10 or want widest compatibility

### glibc231 (Debian 11)
- **Minimum required:** glibc 2.31 or higher
- **Runs on systems with:** glibc ≥2.31 (forward compatible)
- **Compatible distributions:**
  - Debian 11, 12, 13+
  - Ubuntu 20.04, 22.04, 24.04+
  - CentOS 8, Stream 9
  - Fedora 32+
- **Build platform:** Debian 11 (bullseye)
- **Use when:** You're on 2020+ distributions (but not Debian 10)

### glibc236 (Debian 12)
- **Minimum required:** glibc 2.36 or higher
- **Runs on systems with:** glibc ≥2.36 (forward compatible)
- **Compatible distributions:**
  - Debian 12, 13+
  - Ubuntu 22.04, 24.04+
  - RHEL 9, Rocky 9, Alma 9
  - Fedora 37+
- **Build platform:** Debian 12 (bookworm)
- **Use when:** You're on 2023+ distributions

### glibc239 (Ubuntu 24.04)
- **Minimum required:** glibc 2.39 or higher
- **Compatible distributions:**
  - Ubuntu 24.04 LTS (Noble Numbat) and later
  - Fedora 40 and later
  - Debian 13+ (forward compatible)
- **Build platform:** Ubuntu 24.04
- **Use when:** You're on the latest distributions

## How to Check Your glibc Version

```bash
# Method 1: Using ldd
ldd --version | head -1
# Output examples:
# ldd (Debian GLIBC 2.36-5)       ← Your glibc is 2.36
# ldd (Ubuntu GLIBC 2.39-0ubuntu8) ← Your glibc is 2.39

# Method 2: Using libc.so directly
/lib/x86_64-linux-gnu/libc.so.6
# Output: GNU C Library (Debian GLIBC 2.36-5) stable release version 2.36
```

## Quick Selection Guide

| Your glibc Version | Use Binary |
|--------------------|------------|
| 2.28 - 2.30 | `glibc228` |
| 2.31 - 2.35 | `glibc231` |
| 2.36 - 2.38 | `glibc236` |
| 2.39+ | `glibc239` |
| Alpine (musl) | Use static (`*_linux_*.tar.gz`) |
| Not sure | Use static (`*_linux_*.tar.gz`) |

## Distribution to Binary Mapping

| Distribution | Version | Use Binary |
|--------------|---------|------------|
| Debian | 10 (Buster) | `glibc228` |
| Debian | 11 (Bullseye) | `glibc231` |
| Debian | 12 (Bookworm) | `glibc236` |
| Debian | 13 (Trixie) | `glibc239` |
| Ubuntu | 18.04 LTS | **Use static** (glibc 2.27) |
| Ubuntu | 20.04 LTS | `glibc231` |
| Ubuntu | 22.04 LTS | `glibc236` |
| Ubuntu | 24.04 LTS | `glibc239` |
| CentOS | 7 | `glibc228` |
| CentOS | 8 | `glibc231` |
| RHEL / Rocky / Alma | 9 | `glibc236` |
| Alpine | Any | `*_linux_*.tar.gz` (static) |

## Installation Example

```bash
# Check your glibc version first
ldd --version | head -1

# Example output: ldd (Debian GLIBC 2.36-5)
# You have glibc 2.36, use glibc236 binary

# Download and install
wget https://github.com/yourorg/service-user/releases/download/v1.0.0/service-user_1.0.0_linux_glibc236_amd64.tar.gz
tar -xzf service-user_1.0.0_linux_glibc236_amd64.tar.gz
./service-user --version

# Verify dynamic linking
ldd ./service-user
# Output will show: libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
```

## Troubleshooting

### "version 'GLIBC_2.xx' not found"

This error means the binary requires a newer glibc version than your system has.

**Solution:** Either:
1. Use a binary with a lower glibc requirement
2. Use the static binary (works everywhere)
3. Upgrade your system

### "No such file or directory" on execution

This can occur on ARM systems if you downloaded the amd64 binary (or vice versa).

**Solution:** Verify your architecture:
```bash
uname -m
# x86_64 → use amd64 binaries
# aarch64 → use arm64 binaries
```
