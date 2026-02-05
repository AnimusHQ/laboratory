# DevEnv: эксплуатационные требования и контроль

**Версия документа:** 1.0

## Назначение
DevEnv (управляемая среда разработки) предоставляет интерактивный доступ к рабочему окружению, согласованному с политиками проекта. Контрольная плоскость (CP) выпускает краткоживущие сессии доступа, плоскость данных (DP) создаёт и обслуживает вычислительный workload, а политики и аудит фиксируются при каждом изменении состояния.

## Архитектурная схема
- **CP**: создание DevEnv, выпуск сессий доступа, аудит, периодическая сверка TTL.
- **DP**: создание/удаление workload (Kubernetes Job), проверка готовности окружения.
- **Политики**: фиксируются в `PolicySnapshot` при создании DevEnv и неизменны для жизненного цикла среды.

## Потоки
1) **Создание DevEnv**
- Вход: `templateRef`, `ttlSeconds`, `Idempotency-Key`.
- CP строит `PolicySnapshot`, фиксирует сущность DevEnv и снапшот политики.
- CP вызывает DP (provision), DP создаёт Job с ограничениями ресурсов и TTL.
- CP записывает аудит: `devenv.created`, `devenv.provisioned` или `devenv.provision_failed`.

2) **Доступ**
- Вход: `ttlSeconds` для сессии.
- CP выпускает `DevEnvAccessSession` и записывает `devenv.access.issued`.
- CP обращается к DP за статусом доступа (proxy‑проверка готовности), фиксирует `devenv.access.proxy`.

3) **Остановка**
- CP инициирует удаление workload в DP, фиксирует `devenv.deleted`.

4) **TTL‑сверка**
- Фоновый reconciler в CP помечает истёкшие DevEnv как `expired` и инициирует удаление в DP.
- События: `devenv.expired`, `devenv.deleted`.

## Конфигурация
### Control Plane
- `ANIMUS_DEVENV_TTL` — TTL DevEnv по умолчанию (например, `2h`).
- `ANIMUS_DEVENV_ACCESS_TTL` — TTL сессии доступа (например, `15m`).
- `ANIMUS_DEVENV_RECONCILE_INTERVAL` — период TTL‑сверки (например, `30s`).

### Data Plane
- `ANIMUS_DEVENV_K8S_NAMESPACE` — namespace для DevEnv workloads.
- `ANIMUS_DEVENV_K8S_SERVICE_ACCOUNT` — service account для workloads (минимальные привилегии).
- `ANIMUS_DEVENV_JOB_TTL_AFTER_FINISHED` — TTL для завершённых Job (секунды).

## Политики и безопасность
- DevEnv создаётся с фиксированным `PolicySnapshot`; изменения политик не применяются ретроактивно.
- Сеть: применяется deny‑by‑default egress, разрешения — через `NetworkClassRef`.
- Секреты: доступ только через DP, CP не получает значения секретов.

## Аудит и наблюдаемость
- Все операции DevEnv фиксируются в append‑only аудит‑логе.
- Ключевые события: `devenv.created`, `devenv.provisioned`, `devenv.provision_failed`, `devenv.access.issued`, `devenv.access.proxy`, `devenv.expired`, `devenv.deleted`.

## Ограничения
- Все create‑операции требуют `Idempotency-Key` и должны быть повторяемыми без дублирования сущностей.
- DevEnv создаётся только из зарегистрированного `EnvironmentDefinition`.
