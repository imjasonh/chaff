# `chaff`

This tool reports on deleted files in container images.

These files can be included in your container image due to poor build hygiene, for example, by misusing Dockerfiles:

```
FROM base-image
RUN download-large-file.sh > large.zip
RUN unzip large.zip
RUN rm large.zip
```

This Dockerfile will include `large.zip` in your container image layers, even though it won't be available when the image is run.

Large chaff files bloat image sizes, and can even include sensitive data such as secrets.
Consider this example:

```
FROM base-image
RUN download-secret.sh > secret.key
RUN download-artifact.sh --key=secret.key > large.zip
RUN rm secret.key
```

The secret key is still present in the container image's layers!
`chaff` can help you find them.

# Installation

```
go install github.com/imjasonh/chaff@latest
```

# Usage

```
chaff registry.biz/my/container/image:latest
```

# Example

You can build and publish a chaffy image from [`./example/`](./example):

```
docker buildx build --push -t my-image -f example/Dockerfile.chaff example
```

Then run `chaff` on it to see a report about hidden/deleted files:

```
$ chaff my-image
==== CHAFF REPORT ====
- layers: 10
- total chaff files: 219
- total chaff size: 45 MB
--- random.txt (26 MB)
--- var/lib/apt/lists/deb.debian.org_debian_dists_bullseye_main_binary-arm64_Packages.lz4 (17 MB)
--- var/cache/debconf/templates.dat-old (780 kB)
--- var/cache/debconf/templates.dat (780 kB)
--- var/lib/apt/lists/security.debian.org_debian-security_dists_bullseye-security_main_binary-arm64_Packages.lz4 (306 kB)
--- random.txt (257 kB)
--- var/lib/apt/lists/deb.debian.org_debian_dists_bullseye_InRelease (116 kB)
--- var/lib/dpkg/status-old (83 kB)
--- var/lib/dpkg/status (83 kB)
--- var/lib/apt/lists/security.debian.org_debian-security_dists_bullseye-security_InRelease (44 kB)
--- var/lib/apt/lists/deb.debian.org_debian_dists_bullseye-updates_InRelease (39 kB)
--- etc/ld.so.cache (6.3 kB)
--- var/lib/apt/extended_states (5.6 kB)
--- var/cache/debconf/config.dat-old (4.8 kB)
--- var/cache/debconf/config.dat (4.8 kB)
--- var/log/apt/eipp.log.xz (4.7 kB)
--- var/lib/apt/lists/deb.debian.org_debian_dists_bullseye-updates_main_binary-arm64_Packages.lz4 (3.9 kB)
--- random.txt (3.6 kB)
--- secret.key (82 B)
```

You can then rebuild the images without the unnecessary deleted files:

```
docker buildx build --push -t my-image:fixed -f example/Dockerfile.unchaffed example
```

And look for chaff:

```
$ chaff my-image:fixed
==== CHAFF REPORT ====
- layers: 2
- total chaff files: 187
- total chaff size: 1.8 MB
--- var/cache/debconf/templates.dat (780 kB)
--- var/cache/debconf/templates.dat-old (780 kB)
--- var/lib/dpkg/status-old (83 kB)
--- var/lib/dpkg/status (83 kB)
--- etc/ld.so.cache (6.3 kB)
--- var/lib/apt/extended_states (5.6 kB)
--- var/cache/debconf/config.dat-old (4.8 kB)
--- var/cache/debconf/config.dat (4.8 kB)
--- var/log/apt/eipp.log.xz (4.7 kB)
```

These are files from the `debian` base image that your later steps have deleted or overwritten.
