## 1. OpenSpec Change Artifacts

- [ ] 1.1 Create proposal.md (done)
- [ ] 1.2 Create design.md (done)
- [ ] 1.3 Create this tasks.md

## 2. Dockerfiles

- [ ] 2.1 Create `crawler/Dockerfile.surfer` — python:3.13-slim + Poetry, `poetry install --only main --no-root`, CMD `python -m surfer`
- [ ] 2.2 Create `crawler/Dockerfile.parser` — golang:1.26 builder (debian) → scratch final, CGO_ENABLED=0, `/bin/parser` entrypoint
- [ ] 2.3 Create `crawler/.dockerignore` — exclude bin/, coverage*, __pycache__/, tests/, .ruff_cache/, .mypy_cache/, go.work.sum

## 3. Docker Compose Updates

- [ ] 3.1 Update `infra/docker/docker-compose.yml` — change `DEFAULT_NAMESPACE` from full path to `crawler`
- [ ] 3.2 Update `crawler/docker-compose.yml`:
  - [ ] 3.2.1 Add `temporal-network` and `example-site-network` as external networks
  - [ ] 3.2.2 Add `surfer` service (image `crawler-surfer:latest`, build Dockerfile.surfer, networks crawler-network + temporal-network, depends_on migrate)
  - [ ] 3.2.3 Add `downloader` service (image `crawler-surfer:latest`, command `python -m downloader`, networks crawler-network + temporal-network + example-site-network, depends_on migrate)
  - [ ] 3.2.4 Add `parser` service (image `crawler-parser:latest`, build Dockerfile.parser, networks crawler-network + temporal-network, depends_on migrate)

## 4. Verification

- [ ] 4.1 Run `openspec validate --change crawler-docker-services` — expect pass
- [ ] 4.2 `cd infra/docker && docker compose build` — confirm temporal services build
- [ ] 4.3 `cd example-site && docker compose up -d --build` — confirm example-site builds
- [ ] 4.4 `cd crawler && docker compose build` — confirm all three images build without errors
- [ ] 4.5 `cd crawler && docker compose up -d` — confirm all services start
- [ ] 4.6 `docker compose logs surfer downloader parser` — verify workers connect to `temporal:7233`
