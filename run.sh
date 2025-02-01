docker run --rm \
  --cpus=0.5 \
  --memory=128m \
  --memory-swap=256m \
  --pids-limit=50 \
  --read-only \
  --security-opt seccomp=seccomp.json \
  --network=none \
  -v $(pwd)/example:/sandbox \
  runner \
  bash -c '\
      gcc /sandbox/main.c && \
      cat /sandbox/input.txt | timeout 5s /sandbox/a.out && \
      find /sandbox -type f -name "a.out" -delete'
