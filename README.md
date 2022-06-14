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
