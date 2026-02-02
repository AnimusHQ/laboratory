# Матрица критериев production‑grade

**Версия документа:** 1.0

Ниже приведено соответствие критериев production‑grade (AC‑01…AC‑10) конкретным возможностям платформы, контрактам API, тестам и документации.

| Критерий | Реализация (фичи/сущности) | Эндпоинты (OpenAPI) | Тесты/проверки | Документы |
| --- | --- | --- | --- | --- |
| AC‑01 Полный ML‑цикл в рамках Project с DatasetVersion + CodeRef + EnvironmentLock | Project, DatasetVersion, Run, CodeRef, EnvironmentLock, PipelineRun | `dataset-registry.yaml`, `experiments.yaml` | `go test ./...`, интеграционный сценарий M2/M6 | `docs/enterprise/04-domain-model.md`, `docs/enterprise/05-execution-model.md` |
| AC‑02 Воспроизводимость production‑run с явными ограничениями | RunSpec, параметры, PolicySnapshot, reproducibility bundle | `experiments.yaml` (RunSpec + bundle) | unit‑тесты валидаторов, проверка bundle | `docs/enterprise/06-reproducibility-and-determinism.md`, `docs/roadmap/execution-backlog.md` |
| AC‑03 Полный аудит и экспорт | AuditEvent append‑only, экспорт NDJSON/SIEM | `audit.yaml`, `experiments.yaml` | audit regression suite, экспорт | `docs/enterprise/08-security-model.md`, `docs/enterprise/09-operations-and-reliability.md` |
| AC‑04 SSO + RBAC на все операции | OIDC/SAML, RBAC матрица, object‑level checks | `gateway.yaml`, все сервисы | security regression suite | `docs/enterprise/08-rbac-matrix.md` |
| AC‑05 Секреты с TTL и редактированием | Secrets manager, redaction pipeline | `experiments.yaml` (execution), DP протокол | security suite (redaction) | `docs/enterprise/08-security-model.md` |
| AC‑06 Авто‑установка/апгрейд/rollback, air‑gapped | Helm/Kustomize, миграции, SBOM | Helm charts, values schema | upgrade/rollback tests | `docs/enterprise/09-operations-and-reliability.md`, `docs/enterprise/10-versioning-and-compatibility.md` |
| AC‑07 Backup/DR с RPO/RTO | backup tooling, restore drill | N/A | DR game‑day | `docs/enterprise/09-operations-and-reliability.md` |
| AC‑08 Метрики/логи/трейсы CP+DP | Prometheus + OTel, correlation IDs | все сервисы | observability smoke | `docs/enterprise/09-operations-and-reliability.md` |
| AC‑09 DevEnv без обхода политики | DevEnv controller + audit | DP протокол + CP APIs | e2e DevEnv | `docs/enterprise/07-developer-environment.md` |
| AC‑10 Отсутствие скрытого состояния | явные сущности и связи, lineage/audit | `lineage.yaml`, `audit.yaml`, `experiments.yaml` | reproducibility bundle verification | `docs/enterprise/04-domain-model.md`, `docs/enterprise/06-reproducibility-and-determinism.md` |
