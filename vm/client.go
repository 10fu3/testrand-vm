package vm

import (
	"github.com/google/uuid"
	"sync"
	"testrand-vm/compile"
)

func NewSupervisor() *Supervisor {
	return &Supervisor{}
}

type GroupTaskId string
type TaskId string

type Supervisor struct {
	Mutex        *sync.RWMutex
	Tasks        map[TaskId]*Closure
	RefGroupTask map[TaskId]GroupTaskId
	GroupTask    map[GroupTaskId]*struct {
		Count      uint
		Complete   uint
		Tasks      []TaskId
		OnComplete *Closure
	}
}

/**
@sendTask: 送信するタスク
@onComplete: タスクが完了した時に呼び出す関数
@return: タスクID
*/

func (s *Supervisor) AddTaskTaskWithCallback(sendTask compile.SExpression, onComplete *Closure) TaskId {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	taskId := TaskId(uuid.NewString())
	s.Tasks[taskId] = onComplete
	return taskId
}

/*
*
@sendTask: 送信するタスク
@onComplete: タスクが完了した時に呼び出す関数
@return: タスクID
*/
func (s *Supervisor) AddGroupTasks(sendTask []compile.SExpression, onComplete *Closure) GroupTaskId {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	groupId := GroupTaskId(uuid.NewString())
	sendTaskSize := len(sendTask)
	for i := 0; i < sendTaskSize; i++ {
		taskId := TaskId(uuid.NewString())
		s.RefGroupTask[taskId] = groupId
		s.GroupTask[groupId].Tasks = append(s.GroupTask[groupId].Tasks, taskId)
	}
	s.GroupTask[groupId].OnComplete = onComplete
	s.GroupTask[groupId].Count = uint(sendTaskSize)
	return groupId
}

func (s *Supervisor) AddTask(sendTask compile.SExpression) TaskId {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	taskId := TaskId(uuid.NewString())
	return taskId
}

/*
*
@taskId: タスクID
*/
func (s *Supervisor) CompleteTask(taskId string) {
	tId := TaskId(taskId)
	s.Mutex.RLock()
	if groupId, ok := s.RefGroupTask[tId]; ok {
		s.Mutex.RUnlock()
		s.Mutex.Lock()
		s.GroupTask[groupId].Complete++
		if s.GroupTask[groupId].Complete == s.GroupTask[groupId].Count {
			s.Mutex.Unlock()
			s.Mutex.Lock()
			delete(s.GroupTask, groupId)
			for i := 0; i < len(s.GroupTask[groupId].Tasks); i++ {
				delete(s.Tasks, s.GroupTask[groupId].Tasks[i])
			}
			delete(s.RefGroupTask, tId)
			s.Mutex.Unlock()
		}
		s.Mutex.Unlock()
	} else {
		s.Mutex.RUnlock()
		s.Mutex.Lock()
		s.Tasks[tId] = nil
		s.Mutex.Unlock()
	}
}

var superV *Supervisor

func GetSupervisor() *Supervisor {
	return superV
}

func StartSupervisorForClient() *Supervisor {
	supervisor := NewSupervisor()
	supervisor.Mutex = &sync.RWMutex{}
	supervisor.Tasks = make(map[TaskId]*Closure)
	supervisor.RefGroupTask = make(map[TaskId]GroupTaskId)
	supervisor.GroupTask = make(map[GroupTaskId]*struct {
		Count      uint
		Complete   uint
		Tasks      []TaskId
		OnComplete *Closure
	})
	return supervisor
}
