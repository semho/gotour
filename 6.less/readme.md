## Вопросы
1. Различные типы контекстов.
   context.Background()
   context.TODO()
   WithCancel(parent context.Context)
   WithDeadline(parent context.Context, d time.Time)
   WithTimeout(parent context.Context, timeout time.Duration)
   WithValue(parent context.Context, key, val interface{})
2. Graceful Shutdown, Graceful Degradation в микросервисной архитектуре.
   https://habr.com/ru/companies/timeweb/articles/589167/
   https://habr.com/ru/articles/771626/
   https://nuancesprog-ru.turbopages.org/turbo/nuancesprog.ru/s/p/12694/
   https://sky.pro/wiki/javascript/graceful-degradation-v-veb-osnovy-primery-i-realizaciya/
3. Паттерн Health Check API
   https://medium.com/golang-notes/go-http-health-checker-78ce47a36dfa
   https://habr.com/ru/companies/otus/articles/676342/
   https://golang.hotexamples.com/ru/examples/google.golang.org.grpc.health/-/HealthCheck/golang-healthcheck-function-examples.html
4. Ошибки. Работа с ошибками в Go. Паттерны многопоточной обработки ошибок.
   https://habr.com/ru/companies/vk/articles/473658/
   https://tproger.ru/articles/obrabotka-oshibok-v-go
   https://golangify.com/errors
   https://habr.com/ru/companies/otus/articles/558404/
5. errors.Is / errors.As
   https://habr.com/ru/companies/vk/articles/473658/
   https://cs.opensource.google/go/go/+/refs/tags/go1.21.4:src/errors/wrap.go;l=44
6. Паники. Конструкция defer. Функция recover()
   https://go.dev/blog/defer-panic-and-recover
   https://medium.com/nuances-of-programming/%D0%BE%D0%B1%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%BA%D0%B0-%D0%BE%D1%88%D0%B8%D0%B1%D0%BE%D0%BA-%D0%B2-golang-%D1%81-%D0%BF%D0%BE%D0%BC%D0%BE%D1%89%D1%8C%D1%8E-panic-defer-%D0%B8-recover-ea6bfd357af1
   https://folko.gitbook.io/goland/specifikaciya-1/obrabotka-paniki
   ```go
   func protect(g func()) {
      defer func() {
        // Println выполняется нормально, 
        // даже если есть паника
        log.Println("done")  
        if x := recover(); x != nil {
            log.Printf("run time panic: %v", x)
        }
      }()
      log.Println("start")
      g()
   }
    ```
7. Логгирование ошибок. Агрегация ошибок. Sentry.
   https://habr.com/ru/articles/745876/
   https://pkg.go.dev/github.com/muravjov/slog/sentry
   https://sentry-blog.sentry.dev/logging-go-errors/
   https://medium.com/codex/making-go-errors-play-nicely-with-sentry-3ac2cc423cd0
8. Метрики. Работа с Prometheus.
   https://habr.com/ru/companies/otus/articles/769806/
   https://eax.me/golang-prometheus-metrics/

## Практика
### 1. Вычисление Пи
Вычисление числа Пи при помощи ряда Лейбница
https://ru.wikipedia.org/wiki/%D0%A0%D1%8F%D0%B4_%D0%9B%D0%B5%D0%B9%D0%B1%D0%BD%D0%B8%D1%86%D0%B0

Запускаем N горутин (N передаем через флаг)
Каждая горутина вычисляет свою часть ряда.

Например, при N = 1, единственная горутина считает каждый элемент
При N = 2, первая горутина считает сумму всех четных элементов ряда, вторая - сумму всех нечетных элементов.
…

При получении SIGINT или SIGTERM горутины должны завершаться, а на экран выводиться приблизительное значение числа Пи (сумма всех результатов из каждой горутины умноженная на 4)

### 2. Семафор
Сделать 3 различных реализации семафора
```go
type Semaphore interface {
    Acquire(context.Context, int64) error
    TryAcquire(int64) bool
    Release(int64)
}
```
### 3. done-каналы в single-канал
Реализовать функцию, которая будет объединять один или более done-каналов в single-канал, если один из его составляющих каналов закроется, то закроется и он сам.
Очевидным вариантом решения могло бы стать выражение при использованием select, которое бы реализовывало эту связь, однако иногда неизвестно общее число done-каналов, с которыми вы работаете в рантайме. В этом случае удобнее использовать вызов единственной функции, которая, приняв на вход один или более or-каналов, реализовывала бы весь функционал.

Определение функции:
```go
var or func(channels ...<- chan interface{}) <- chan interface{}
```

Пример использования функции:

```go
sig := func(after time.Duration) <- chan interface{} {
    c := make(chan interface{})
    go func() {
        defer close(c)
        time.Sleep(after)
    }()
    return c
}

start := time.Now()
<-or (
    sig(2*time.Hour),
    sig(5*time.Minute),
    sig(1*time.Second),
    sig(1*time.Hour),
    sig(1*time.Minute),
)

fmt.Printf(“done after %v”, time.Since(start)) // ~1 second
```
