# runner

This is an image for `runner` container. Some sort of a sandbox, inside which users' solutions are executed

### Deploy

Build an image with initial tag

```bash
$ docker build -t runner .
```

Create tag for container registry

```bash
$ docker tag runner:latest ghcr.io/voidcontests/runner:latest
```

Push an image to container registry

```bash
$ docker push ghcr.io/voidcontests/runner:latest
```
