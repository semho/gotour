## Вопросы
1. Горутины. https://go.ptflp.ru/course2/8/8.1.1/
2. Внутреннее устройство горутин. https://habr.com/ru/companies/otus/articles/527748/
3. Планировщик Go. https://www.youtube.com/watch?v=P2Tzdg8n9hw&t=3154s
4. Go Runtime и его составляющие. https://go.ptflp.ru/course2/9/9.4.1/ https://runebook.dev/ru/docs/go/runtime/index
5. Аллокация в стеке и в куче. https://dzen.ru/a/ZT_hVKNwlSdgtNPG https://folko.gitbook.io/goland/optimizaciya/untitled
6. Escape Analysis.  https://habr.com/ru/articles/497994/ https://habr.com/ru/companies/oleg-bunin/articles/676332/  
   ##### У вас 100% выделится значение на хипе, если:
   * Возврат результата происходит по ссылке;
   * Значение передается в аргумент типа interface{} — аргумент fmt.Println;
   * Размер значения переменной превышает лимиты стека.  
   * ```
     type X struct {
       p *int
     }
     var i1 int //заполним числом
     x1 := &X{
        p: &i1      // в стэк
     }
     
     var i2 int  //заполним числом
     x2 := &X{}
     x2.p = &i2  // а тут уже в кучу
     ```
7. Garbage Collector.  https://habr.com/ru/companies/avito/articles/753244/ https://www.youtube.com/watch?v=ZZJBu2o-NBU
8. Каналы. https://habr.com/ru/articles/490336/
9. Внутреннее устройство каналов. https://medium.com/@victor_nerd/%D0%BF%D0%BE%D0%B4-%D0%BA%D0%B0%D0%BF%D0%BE%D1%82%D0%BE%D0%BC-golang-%D0%BA%D0%B0%D0%BA-%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%B0%D1%8E%D1%82-%D0%BA%D0%B0%D0%BD%D0%B0%D0%BB%D1%8B-%D1%87%D0%B0%D1%81%D1%82%D1%8C-1-e1da9e3e104d https://medium.com/@victor_nerd/golang-channel-internal-part2-b4e37ad9a118
10. Аксиомы каналов https://dzen.ru/a/ZT37Gzpya2uEvz9L

## Практика
Реализовать паттерн "Worker Pool"
Необходимо написать функцию для параллельного выполнения заданий в N параллельных горутинах:

* количество создаваемых горутин не должно зависеть от числа заданий, т.е. функция должна запускать N горутин для конкурентной обработки заданий и, возможно, еще несколько вспомогательных горутин;
* функция должна останавливать свою работу, если произошло m ошибок;
* после завершения работы функции (успешного или из-за превышения M) не должно оставаться работающих горутин.
* Нужно учесть, что задания могут выполняться разное время, а длина списка задач len(tasks) может быть больше или меньше N.

Значение M <= 0 трактуется на усмотрение программиста:
* или это знак игнорировать ошибки в принципе;
* или считать это как "максимум 0 ошибок", значит функция всегда будет возвращать ErrErrorsLimitExceeded
на эту логику следует написать юнит-тест.

#### Граничные случаи

* если задачи работают без ошибок, то выполнятся len(tasks) задач, т.е. все задачи;
* если в первых выполненных M задачах (или вообще всех) происходят ошибки, то всего выполнится не более N+M задач.

#### (*) Дополнительное задание: написать тест на concurrency без time.Sleep
```
import (
"errors"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
    // Place your code here.
    return nil
}
```