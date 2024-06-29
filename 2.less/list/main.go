package main

import "fmt"

type Node struct {
	val  int
	next *Node
}

func (n *Node) Push(node *Node, val int) *Node {
	if node == nil {
		return &Node{val: val}
	}

	current := node
	for current.next != nil {
		current = node.next
	}
	current.next = &Node{val: val}

	return node
}

func Print(node *Node) {
	current := node
	for current != nil {
		fmt.Printf("%d -> ", current.val)
		current = current.next
	}
	fmt.Print(nil)
}

func main() {
	var head *Node

	head = head.Push(head, 1)
	head = head.Push(head, 5)

	Print(head)
}
