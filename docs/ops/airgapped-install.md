# Air‑gapped установка (offline)

Документ описывает подготовку офлайн‑набора и установку без доступа к интернету.

## 1. Подготовка образов
1. Зафиксируйте дигесты в файлах значений (пример):
```yaml
image:
  repository: registry.local/animus
  tag: "0.1.0"
  digest: "sha256:..." # или image.digests для отдельных сервисов
```
2. Получите список образов:
```bash
scripts/list_images.sh --values values-control-plane.yaml --values values-data-plane.yaml > images.txt
```
3. Загрузите образы в локальный registry и убедитесь, что они доступны по digests.

## 2. Формирование offline‑набора
Используйте пакетировщик:
```bash
closed/scripts/airgap-bundle.sh --output ./bundle --values values-control-plane.yaml --values values-data-plane.yaml
```
Внутри `bundle/` будет:
- Helm‑чарты (`*.tgz`)
- `images.txt` и `images.tar`
- `SHA256SUMS`

## 3. Проверка целостности
```bash
cd bundle
sha256sum -c SHA256SUMS
```

## 4. Развёртывание в изолированной среде
1. Импортируйте образы:
```bash
docker load -i images.tar
```
2. Передайте образы в registry, доступный кластеру.
3. Установите чарты:
```bash
helm upgrade --install animus-cp animus-datapilot-*.tgz -f values-control-plane.yaml
helm upgrade --install animus-dp animus-dataplane-*.tgz -f values-data-plane.yaml
```

## 5. Примечания
- Все образы должны быть digest‑пиннены для воспроизводимости.
- Для внешних Postgres/S3 используйте значения `database.url` и `minio.*`.
