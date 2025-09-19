package userserver

import (
	"context"
	"myGinServer/internal/store"
	"myGinServer/models/task"
)

type TasksServer struct {
	dbStore store.DBStore
}

func NewTasksServer(dbStore store.DBStore) *TasksServer {
	return &TasksServer{dbStore: dbStore}
}

func (s *TasksServer) TaskList(ctx context.Context, keyword string) ([]task.Task, error) {
	return s.dbStore.ListTasks(ctx, keyword)
}

func (s *TasksServer) TaskDel(ctx context.Context, id string) error {
	return s.dbStore.DelTasks(ctx, id)
}
func (s *TasksServer) TaskSave(ctx context.Context, task task.Task) (string, error) {
	return s.dbStore.SaveTask(ctx, task)
}
