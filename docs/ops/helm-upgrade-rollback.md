# Helm‑апгрейд и rollback

## Апгрейд
1. Сформируйте новые значения (`values-control-plane.yaml`, `values-data-plane.yaml`) и убедитесь, что образы доступны.
2. Выполните апгрейд:
```bash
helm upgrade animus-cp ./closed/deploy/helm/animus-datapilot -f values-control-plane.yaml
helm upgrade animus-dp ./closed/deploy/helm/animus-dataplane -f values-data-plane.yaml
```
3. Проверьте статус:
```bash
helm status animus-cp
helm status animus-dp
```

Важно: migrations‑job запускается как post‑install/post‑upgrade hook. При несовместимых изменениях схемы БД обязательно выполните резервное копирование.

## Rollback
1. Посмотрите историю релизов:
```bash
helm history animus-cp
helm history animus-dp
```
2. Выполните откат:
```bash
helm rollback animus-cp <revision>
helm rollback animus-dp <revision>
```
3. Повторно проверьте readiness и базовые API.

## Рекомендации
- Держите резервные копии Postgres и MinIO/S3 перед каждым апгрейдом.
- Для upgrade с несовместимыми миграциями используйте отдельное окно обслуживания.
