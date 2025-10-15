package etcd

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	nodeId    = uuid.New().String()
	scheduler *TaskScheduler
)

func init() {
	scheduler = NewTaskScheduler([]string{"localhost:2379"}, nodeId)
}

func TestStart(t *testing.T) {

	err := scheduler.Start()
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	defer scheduler.Stop()
}
