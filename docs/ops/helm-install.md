# Helm‑установка Control Plane и Data Plane

Документ описывает базовую установку компонентов Animus в Kubernetes через Helm.

## 1. Предусловия
- Kubernetes 1.25+.
- Helm 3.7+ (поддержка values.schema.json).
- Доступные образы в реестре (tag или digest).
- Подготовленные секреты для `ANIMUS_INTERNAL_AUTH_SECRET` и CI webhook.

## 2. Подготовка значений
Создайте два файла значений:
- `values-control-plane.yaml` для `animus-datapilot`.
- `values-data-plane.yaml` для `animus-dataplane`.

Минимальный пример (control plane):
```yaml
image:
  repository: ghcr.io/animus-labs
  tag: "0.1.0"

auth:
  mode: dev
  internalAuthSecret: "<shared-secret>"

ci:
  webhookSecret: "<ci-secret>"

database:
  url: "postgres://animus:animus@postgres.example:5432/animus?sslmode=disable"

postgres:
  enabled: false

minio:
  enabled: false
  endpoint: "s3.example.local:9000"
  accessKey: "<access-key>"
  secretKey: "<secret-key>"
  region: us-east-1
  useSSL: false
  buckets:
    datasets: datasets
    artifacts: artifacts
```

Минимальный пример (data plane):
```yaml
image:
  repository: ghcr.io/animus-labs
  tag: "0.1.0"

auth:
  internalAuthSecret: "<shared-secret>"

controlPlane:
  baseURL: "http://animus-cp-gateway:8080"
```

## 3. Установка
```bash
helm upgrade --install animus-cp ./closed/deploy/helm/animus-datapilot -f values-control-plane.yaml
helm upgrade --install animus-dp ./closed/deploy/helm/animus-dataplane -f values-data-plane.yaml
```

## 4. Валидация
- `helm test animus-cp`
- `helm test animus-dp`
- `kubectl get pods` и проверка `/readyz`.

## 5. Внешние Postgres и S3
- Для внешней БД: задайте `database.url`, установите `postgres.enabled=false`.
- Для внешнего S3/MinIO: задайте `minio.enabled=false`, заполните `minio.endpoint`, `minio.region`, `minio.accessKey`, `minio.secretKey`.

## 6. Дигест‑пиннинг
Для air‑gapped установок рекомендуется фиксировать `image.digest` или `image.digests` в `values-control-plane.yaml`.
