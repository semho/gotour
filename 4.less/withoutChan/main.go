package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task struct {
	time time.Duration
	id   int
}

func newTask(args ...time.Duration) []*Task {
	tasks := make([]*Task, len(args))

	for i, dur := range args {
		tasks[i] = &Task{id: i + 1, time: dur}
	}

	return tasks
}

type TaskManager struct {
	tasks             []*Task
	numGoroutines     int
	tasksByGoroutines [][]*Task
	wg                *sync.WaitGroup
	maxError          int
	countError        int
	mu                *sync.RWMutex
	stopWork          bool
}

func newTaskManager(tasks []*Task, countG, maxErr int) *TaskManager {
	return &TaskManager{
		tasks:         tasks,
		numGoroutines: countG,
		maxError:      maxErr,
		wg:            &sync.WaitGroup{},
		mu:            &sync.RWMutex{},
	}
}

func (tm *TaskManager) distributeTasks() {
	if len(tm.tasks) == 0 {
		fmt.Println("Нет задач для распределения.")
		return
	}

	if tm.numGoroutines <= 0 {
		fmt.Println("Количество горутин должно быть больше 0.")
		return
	}

	tm.tasksByGoroutines = make([][]*Task, tm.numGoroutines)

	for i, task := range tm.tasks {
		gIndex := i % tm.numGoroutines
		tm.tasksByGoroutines[gIndex] = append(tm.tasksByGoroutines[gIndex], task)
	}
}

func (tm *TaskManager) executeTasks() {
	for i, tasksGoroutine := range tm.tasksByGoroutines {
		tm.wg.Add(1)
		go tm.processTasks(i+1, tasksGoroutine)
	}
	tm.wg.Wait()
	fmt.Printf("Всего ошибок: %d\n", tm.countError)
}

func (tm *TaskManager) processTasks(gID int, tasks []*Task) {
	defer tm.wg.Done()
	for _, task := range tasks {
		if tm.checkErrors() { //TODO: не работает
			fmt.Println("Время прервать выполнение задач")
		}

		tm.wg.Add(1)
		go func(gID int, task *Task) {
			defer tm.wg.Done()
			err := tm.someWorkTask(gID, task)
			if err != nil {
				tm.handleError(err)
			}
		}(gID, task)
	}
}

func (tm *TaskManager) someWorkTask(gID int, task *Task) error {
	time.Sleep(task.time)
	if task.id == 1 {
		return ErrErrorsLimitExceeded
	}
	fmt.Printf(
		"Горутина %d: Таска %d завершила свою работу за %d секунд\n",
		gID,
		task.id,
		int64(task.time/time.Second),
	)
	return nil
}

func (tm *TaskManager) handleError(err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.countError++
	fmt.Println(err)
	if tm.countError >= tm.maxError {
		tm.stopWork = true
	}
}

func (tm *TaskManager) checkErrors() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.stopWork
}

func main() {
	tasks := newTask(
		1*time.Second,
		3*time.Second,
		2*time.Second,
		1*time.Second,
		7*time.Second,
		5*time.Second,
		1*time.Second,
	)

	taskManager := newTaskManager(tasks, 2, 1)

	taskManager.distributeTasks()
	taskManager.executeTasks()
}
