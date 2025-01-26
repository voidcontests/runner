docker run --rm \
  --cpus=0.5 \
  --memory=128m \
  --memory-swap=256m \
  --pids-limit=50 \
  --read-only \
  -v $(pwd):/sandbox \
  runner \
  bash -c '\
      gcc /sandbox/main.c && \
      cat /sandbox/input.txt | /sandbox/a.out > /sandbox/sample.txt && \
      if diff -q /sandbox/sample.txt /sandbox/output.txt > /dev/null; then \
        echo "OK"; \
      else \
        echo "WA"; \
      fi'
