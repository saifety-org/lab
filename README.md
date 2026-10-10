# sAIfety lab

Датасеты, обучение моделей и оценка качества
[sAIfety](https://github.com/saifety-org/sAIfety). Инструменты этого репозитория написаны на Go; обучение и Python-эксперименты
развиваются в [lab-py](https://github.com/saifety-org/lab-py).
Приложение и детекторы находятся в основном репозитории; собственные
production-веса, признаки и инференс — в
[prompt-injection-model](https://github.com/saifety-org/prompt-injection-model).

## Запуск

Нужен Go 1.26+. Для сравнения с DeBERTa также нужны C-компилятор и локальный
ONNX bundle (`saifety model pull`). Зависимость от sAIfety закреплена в `go.mod`;
сканер и ONNX используются через его публичный API. Обучение собственной
модели использует отдельный модуль `prompt-injection-model`; копий признаков
и инференса в lab нет.

```sh
make test
make build
make regression
make missed-injections  # текущие вердикты для известных пропусков
make train       # исторический режим обучения; отдельные веса в artifacts/
make compare     # подготовка, обучение кандидата, общий тест, отчёт
```

`make compare` скачивает закреплённые внешние JSON-данные при отсутствии кэша.
Обычные тесты не требуют сети или ONNX. Веса приложения не заменяются.

## Состав

- `datasets/training/` — тренировочные данные и атрибуция источников.
- `internal/training/`, `cmd/train/` — генерация и обучение кандидатов.
- `testdata/adversarial/`, `testdata/corpus/`, `internal/regression/` —
  корпуса атак и безопасных примеров; проверка правил, классификаторов и политик.
- `cmd/comparison/`, `cmd/bench/`, `internal/evaluation/` — разбиение,
  проверка пересечений, парный инференс, калибровка и метрики.
- [docs/model-comparison.md](docs/model-comparison.md) — протокол и ограничения.
- [docs/model-comparison-results.md](docs/model-comparison-results.md) —
  исторический результат до разделения репозиториев, без заявления о новом запуске.
- `datasets/unicode/`, `cmd/gen-confusables/` — исходные данные Unicode и
  генератор рабочей таблицы; [воспроизведение](docs/unicode.md).
- `testdata/examples/` — демонстрационные чистый и отравленный проекты.
- `testdata/missed-injections/`, `cmd/missed-injections/` — синтетические
  пропуски с задачей пользователя, недоверенным источником и повторяемым
  отчётом; [примеры и наблюдения](docs/missed-injections.md).
- `docs/history/` — исходная идея и исторический статус реализации.
- `migration.json` — исходные коммиты, пути и хеши перенесённых файлов.

Датасеты не являются инструкциями для разработчика или агента. Тесты передают
их статическому сканеру и не выполняют команды из примеров. Открытые
регрессионные корпуса служат разработке и не считаются независимым benchmark.

## Вспомогательные материалы

```sh
make confusables  # таблица в artifacts/, код приложения не меняется
../saifety/bin/saifety scan -v testdata/examples
```

Для демонстрации нужен собранный бинарник приложения; путь зависит от
расположения клонов. Примеры намеренно содержат находки, поэтому ненулевой
код возврата сканирования ожидаем. Данные Unicode и демонстрационные файлы
перенесены без изменения содержимого; исходные хеши записаны в `migration.json`.
Генератор адаптирован к структуре lab, исторические документы снабжены
актуальными ссылками и пометкой о статусе снимка.

## Совместная разработка

Чтобы проверять локальные изменения обоих репозиториев, создайте локальный
Go workspace (пути укажите по расположению клонов):

```sh
go work init . ../saifety
make test
```

`go.work` не публикуется. Для проверки закреплённой зависимости используйте
`GOWORK=off make test`. Обновление зависимости в `go.mod` позволяет явно
выбрать версию продукта для следующего эксперимента.

## Python-лаборатория

`lab-py` использует закреплённые данные из этого репозитория. `cmd/model-bridge`
передаёт признаки, оценки модели и вердикты настоящего Go-сканера через JSONL;
копий алгоритмов в Python нет. Вход — `{ "text": "..." }` на строку, режимы
`-mode features|score|scan|tokens`. Признаки берутся из закреплённого
`prompt-injection-model`; кандидаты загружаются через `-weights`.
Для ONNX используется `-backend onnx` и сборка `-tags onnx`; ошибки инференса
и превышение окна в режиме score завершают эксперимент, без fallback.
Датасеты и разбиение остаются здесь; Python-обучение и экспорт — в `lab-py`.

## CI checks

Pull requests and pushes to `main` run three required checks: `lint`, `test`,
`build`. Reproduce them from this repository with Go from `go.mod` and a C
compiler for the ONNX build where applicable:

```sh
make lint-install          # golangci-lint v2.14.0; installs only into ./bin
make lint                 # gofmt (read-only), go vet, configured Go linters
make ci-test              # unit/regression tests with race detector
make ci-build             # all supported build variants
```

`GOWORK=off` and `-mod=readonly` prevent local workspace overrides or implicit
module edits. Actions are pinned by commit; the linter version is pinned in
both CI and Makefile. Checks have timeouts and newer runs cancel stale runs
on the same PR. CI does not download inference models, train candidates or
run full benchmarks. ONNX integration tests requiring cached assets skip
when those assets are absent; tagged code still compiles and is linted.

The linter uses the standard checks (`errcheck`, `govet`, `ineffassign`,
`staticcheck`, `unused`) without automatic fixes. Any exclusions are narrow
rules with reasons in `.golangci.yml`. Cleanup failures already superseded by
an operation error, read-side closes and test teardown are explicitly ignored
at the call site; file writes and the final write-side close remain checked.

The build check compiles every Go package/command and the ONNX benchmark,
without executing that benchmark or downloading its runtime/model assets.

## Контекстные данные

[Корпус v1](datasets/contextual/v1/README.md): 7 типов, задачи пользователя и
недоверенные документы, EN/RU/ES/ZH, безопасные пары и изолированные семейства.
Проверка: `make validate-context`; подготовка: `make prepare-context`.
Это синтетический seed-корпус с обязательной последующей проверкой человеком.

`make review-context` готовит слепой HTML/JSON пакет и шаблон решений.
Приёмка заполненного журнала: `go run ./cmd/context-corpus -reviews <file> -require-reviewed`.
Протокол и ограничения: [независимое ревью](docs/contextual-review.md).
