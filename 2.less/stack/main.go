package main

import "fmt"

type Stack []int

func (s *Stack) Push(x int) {
	*s = append(*s, x)
}

func (s *Stack) Pop() (int, error) {
	if len(*s) == 0 {
		return 0, fmt.Errorf("empty queue")
	}

	x := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return x, nil
}

type Stack2 struct {
	stack []int
}

func (s *Stack2) Push(x int) {
	s.stack = append(s.stack, x)
}

func (s *Stack2) Pop() (int, error) {
	if len(s.stack) == 0 {
		return 0, fmt.Errorf("empty queue")
	}

	x := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return x, nil
}

func main() {
	//stack := []int{1}
	//stack = append(stack, 2)
	//stack = append(stack, 3)
	//for _, value := range stack {
	//	fmt.Println(value)
	//}
	//for len(stack) > 0 {
	//	element := stack[len(stack)-1]
	//	stack = stack[:len(stack)-1]
	//	fmt.Println(element)
	//}
	//
	//fmt.Println(stack)

	//stack := Stack{}
	//stack.Push(1)
	//stack.Push(2)
	//
	//for _, value := range stack {
	//	fmt.Println(value)
	//}
	//
	//for len(stack) > 0 {
	//	element, _ := stack.Pop()
	//	fmt.Println(element)
	//}

	stack := Stack2{}
	stack.Push(1)
	stack.Push(2)

	for _, value := range stack.stack {
		fmt.Println(value)
	}

	for len(stack.stack) > 0 {
		element, _ := stack.Pop()
		fmt.Println(element)
	}
}
