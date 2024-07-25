## Вопросы
1. Каналы. https://habr.com/ru/articles/490336/
2. Внутреннее устройство каналов. https://medium.com/@victor_nerd/%D0%BF%D0%BE%D0%B4-%D0%BA%D0%B0%D0%BF%D0%BE%D1%82%D0%BE%D0%BC-golang-%D0%BA%D0%B0%D0%BA-%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%B0%D1%8E%D1%82-%D0%BA%D0%B0%D0%BD%D0%B0%D0%BB%D1%8B-%D1%87%D0%B0%D1%81%D1%82%D1%8C-1-e1da9e3e104d https://medium.com/@victor_nerd/golang-channel-internal-part2-b4e37ad9a118    
```
type hchan struct {
   qcount   uint           // Общее количество элементов данных в очереди канала в данный момент.
   dataqsiz uint           // Размер круговой очереди канала(буфер канала). Это максимальное количество элементов, которое может храниться в буфере канала
   buf      unsafe.Pointer // Указатель на массив, который служит буфером канала. Этот массив содержит dataqsiz элементов.
   elemsize uint16         // Размер одного элемента данных в байтах
   closed   uint32         // Флаг, указывающий, закрыт ли канал (1 - закрыт, 0 - открыт)
   elemtype *_type         // Указатель на информацию о типе элементов, передаваемых через канал
   sendx    uint           // Индекс в круговом буфере, куда будет записан следующий отправленный элемент
   recvx    uint           // Индекс в круговом буфере, откуда будет прочитан следующий полученный элемент
   recvq    waitq          // Список горутин, ожидающих получения данных из канала
   sendq    waitq          // Список горутин, ожидающих возможности отправить данные в канал

   lock mutex              // Мьютекс для защиты всех полей структуры hchan, а также некоторых полей в структурах sudog, заблокированных на этом канале
}
```
3. Аксиомы каналов https://dzen.ru/a/ZT37Gzpya2uEvz9L    
```
| Операция       | nil channel | closed channel |
| -------------- | ----------- | -------------- |
| close(c)       | panic       | panic          |
| read val: <- c | block       | default value  |
| write c <- val | block       | panic          |
```
5. Мультиплексирование https://katcipis.github.io/blog/mux-channels-go/   
это процесс объединения нескольких каналов в один   
```
func multiplex(inputs ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup
    wg.Add(len(inputs))

    for _, input := range inputs {
        go func(ch <-chan int) {
            defer wg.Done()
            for value := range ch {
                output <- value
            }
        }(input)
    }

    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}
```
5. Конструкция Select https://habr.com/ru/articles/490336/ https://www.programiz.com/golang/select
6. Коммуникация и синхронизация горутин  https://www.alldevstack.com/ru/golang/sync.html
7. Пакет sync https://dzen.ru/a/ZIAmWGI5PUidHodE
8. Контексты https://habr.com/ru/companies/nixys/articles/461723/ https://blog.ildarkarymov.ru/posts/context-guide/
9. Работа с контекстами https://stepik.org/lesson/748822/step/1?unit=750663

## Практика
### Concurrently Pipeline
https://golang-blog.blogspot.com/2019/10/concurrency-patterns-pipelines.html
https://www.youtube.com/watch?v=8Rn8yOQH62k

Необходимо реализовать функцию для запуска конкуррентного пайплайна, состоящего из стейджей.

Стейдж - функция, принимающая канал на чтение и отдающая канал на чтение, внутри в горутине берущая данные из входного канала, выполняющая полезную работу и отдающая результат в выходной канал:
```
func Stage(in <-chan interface{}) (out <-chan interface{}) {
out = make(chan interface{})
go func() { /* Some work */ }()
return out
}
```

Особенность пайплайна в том, что обработка последующего элемента входных данных должна происходить без ожидания завершения всего пайплайна для текущего элемента.

Т.е. пайплан из 4 функций по 100 мс каждая для 5 входных элементов должен выполняться гораздо быстрее, чем за 2 секунды (4 * 100 мс * 5).

Также должна быть реализована возможность остановить пайплайн через дополнительный сигнальный канал (done/terminate/etc.).

При необходимости можно выделять дополнительные функции.

Нельзя менять сигнатуры исходных функций.

Учесть, что в функции stage может случиться паника.
```
type (
In  = <-chan interface{}
Out = In
Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
// Place your code here.
return nil
}
```
