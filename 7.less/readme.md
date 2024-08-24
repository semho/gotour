## Вопросы
1. Паттерны многопоточной обработки ошибок. Работа с multierr / errgroup.
   https://dev.to/ryoyamaoka/multiple-error-handling-in-go-3930
2. Логгирование ошибок. Агрегация ошибок. Sentry.
   https://rollbar.com/blog/golang-error-logging-guide/   
   https://golang.withcodeexample.com/blog/mastering-error-handling-logging-go-guide/
3. Метрики. Типы метрик в Prometheus.
   https://prometheus.io/docs/concepts/metric_types/
   https://prometheus.io/docs/tutorials/understanding_metric_types/
   https://habr.com/ru/companies/tochka/articles/685636/
4. Трассировка. Спецификация OpenTelemetry.
   https://habr.com/ru/articles/710644/
   https://opentelemetry.io/docs/languages/go/instrumentation/
   https://opentelemetry.io/docs/concepts/signals/traces/
   https://opentelemetry.io/docs/specs/otel/trace/api/
5. Observability сервисов. SLA / SLO / SLI.
   https://www.atlassian.com/ru/incident-management/kpis/sla-vs-slo-vs-sli
   https://etogeek.dev/posts/sli-slo-sla/
   https://teletype.in/@flow_russian/6fZTECzIinx
   https://habr.com/ru/companies/otus/articles/676342/

## Практика
Реализовать Merge Sort нескольких файлов.

Массив файлов для чтения и имя результирующего файла передается в аргументах

Считаем, что в исходных файлах строки уже отсортированы

Файлы содержат числа в отдельных строка, в случае, если в строке содержится не число - пропускаем такую строку.

Предусмотреть корректные уровни ошибок для различных ситуаций.

Ошибки в том числе должны логгироваться в https://sentry.io/ (там есть бесплатный тарифный план)

Строки, которые не получилось обработать - фиксируем в отдельном файле.

При получении SIGINT / SIGTERM программа должна завершаться, сохранив в результирующий файл то, что удалось отсортировать и распечатать количество обработанных строк из каждого файла

В процессе работы сервис в качестве метрик пишет сколько в данный момент обработано строк, сколько строк с ошибками, сколько файлов в данный момент открыто


Команда для запуска задания: go run ./main.go -inputs a,b -log-level debug