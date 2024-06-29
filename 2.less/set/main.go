package main

import (
	"fmt"
	"sort"
)

type Set struct {
	items map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		items: make(map[string]struct{}),
	}
}

func (s *Set) Add(item string) {
	s.items[item] = struct{}{}
}

func (s *Set) Remove(item string) {
	delete(s.items, item)
}

func (s *Set) Has(item string) bool {
	_, ok := s.items[item]
	return ok
}

func (s *Set) Len() int {
	return len(s.items)
}

func (s *Set) getItems() []string {
	items := make([]string, 0, len(s.items))
	for item := range s.items {
		items = append(items, item)
	}
	sort.Strings(items)
	return items
}

func (s *Set) String() string {
	return fmt.Sprintf("%v", s.getItems())
}

func (s *Set) Union(other *Set) *Set {
	for item := range other.items {
		s.Add(item)
	}

	return s
}

func (s *Set) Intersection(other *Set) *Set {
	setIntersect := NewSet()

	for item := range other.items {
		if s.Has(item) {
			setIntersect.Add(item)
		}
	}

	return setIntersect
}

func (s *Set) Difference(other *Set) *Set {
	setDifference := NewSet()
	for item := range s.items {
		if !other.Has(item) {
			setDifference.Add(item)
		}
	}

	return setDifference
}

func (s *Set) IsSubset(other *Set) bool {
	for item := range s.items {
		if !other.Has(item) {
			return false
		}
	}

	return true
}

func (s *Set) IsSuperset(other *Set) bool {
	return other.IsSubset(s)
}

func main() {
	set := NewSet()
	set.Add("a")
	set.Add("b")
	set.Add("c")

	fmt.Printf("Длина множества: %d\n", set.Len())
	fmt.Printf("Проверка наличность элементов множества: %t\n", set.Has("a"))
	set.Remove("a")
	fmt.Printf("Проверка наличность элементов множества: %t\n", set.Has("a"))
	fmt.Printf("Вывод множества: %s\n", set)

	set2 := NewSet()
	//set2.Add("a")
	set2.Add("b")
	//set2.Add("dd")
	fmt.Printf("Вывод множества2: %s\n", set2)
	//set.Union(set2)
	//fmt.Println(set)

	//res := set.Intersection(set2)
	//fmt.Println(res)

	//dif := set.Difference(set2)
	//fmt.Println(dif)

	//fmt.Print(set2.IsSubset(set))
	fmt.Print(set.IsSuperset(set2))

}
