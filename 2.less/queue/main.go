package main

import "fmt"

type Queue []int

func (q *Queue) Enqueue(value int) {
	*q = append(*q, value)
}

func (q *Queue) Dequeue() (int, error) {
	if len(*q) == 0 {
		return 0, fmt.Errorf("empty queue")
	}
	element := (*q)[0]
	*q = (*q)[1:]
	return element, nil
}

type Queue2 struct {
	queue []int
}

func (q *Queue2) Add(value int) {
	q.queue = append(q.queue, value)
}

func (q *Queue2) Remove() (int, error) {
	if len(q.queue) == 0 || q.queue == nil {
		return 0, fmt.Errorf("empty queue")
	}
	element := q.queue[0]
	q.queue = q.queue[1:]
	return element, nil
}

func main() {
	// Создание очереди без отдельного типа на слайсах
	//a := []string{"a", "b", "c"}
	//for _, value := range a {
	//	fmt.Println(value)
	//}
	//a = append(a, "d")
	//for _, value := range a {
	//	fmt.Println(value)
	//}
	//first := a[0]
	//a = a[1:]
	//fmt.Println("---")
	//fmt.Println(first)
	//for _, value := range a {
	//	fmt.Println(value)
	//}

	// Создание очереди с типом на слайсах
	//queue := Queue{}
	//
	//// Добавление элементов в очередь
	//queue.Enqueue(1)
	//queue.Enqueue(2)
	//queue.Enqueue(3)
	//
	//// Удалить и распечатать каждый элемент
	//for len(queue) > 0 {
	//	element, err := queue.Dequeue()
	//	if err != nil {
	//		fmt.Println(err)
	//		break
	//	}
	//	fmt.Println(element)
	//}

	//Создание очереди на структуре со слайсами
	queue2 := Queue2{}
	queue2.Add(1)
	queue2.Add(2)
	queue2.Add(3)
	if len(queue2.queue) > 0 {
		for _, value := range queue2.queue {
			fmt.Println(value)
		}
	}
	fmt.Println("---")
	for len(queue2.queue) > 0 {
		element, err := queue2.Remove()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(element)
	}

}
