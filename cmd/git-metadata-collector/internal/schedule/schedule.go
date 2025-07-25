package schedule

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

var FetchSize = 200
var FetchThreshold = 30

const IdleInterval = 30 * time.Second

func SetFetchOptions(fetchSize, fetchThreshold int) {
	FetchSize = fetchSize
	FetchThreshold = fetchThreshold
}

var tasks []string = make([]string, 0, FetchSize)
var tasksSet map[string]struct{} = make(map[string]struct{})

var manualTasks []string = make([]string, 0)
var muTasks sync.Mutex
var fetchInProgress bool
var taskCond = sync.NewCond(&muTasks)

// When such thing occur: <1> is thread id
// 1. <1> running a task
// 2. <2> fetching database
// 3. <1> update database and delete task from tasksSet
// 4. <2> append to tasks and tasksSet
// Because things will change when fetching database,
// so we delay real delete after step 4
//
// But another problem appears:
// Until next fetch database, task cannot be start manually.
var (
	toDeleteTasks []string
	muToDelete    sync.Mutex
)

var isStop = false
var muStop sync.Mutex
var stopCond = sync.NewCond(&muStop)

func fetchTasksFromDatabase() error {
	ctx := storage.GetDefaultAppDatabaseContext()
	r := repository.NewRankedGitTaskRepository(ctx)

	for {
		ok := false

		result, err := r.Query(FetchSize)
		if err != nil {
			return err
		}

		muTasks.Lock()
		for r := range result {
			ok = true
			if _, exist := tasksSet[*r.GitLink]; !exist {
				tasks = append(tasks, *r.GitLink)
				tasksSet[*r.GitLink] = struct{}{}
			}
		}

		muToDelete.Lock()
		// real delete tasksSet
		for _, t := range toDeleteTasks {
			delete(tasksSet, t)
		}
		toDeleteTasks = make([]string, 0)
		muToDelete.Unlock()

		muTasks.Unlock()
		if !ok {
			time.Sleep(IdleInterval)
		} else {
			break
		}
	}
	return nil
}

func AddManualTask(task string) {
	muTasks.Lock()
	defer muTasks.Unlock()
	if _, ok := tasksSet[task]; !ok {
		manualTasks = append(manualTasks, task)
		tasksSet[task] = struct{}{}
	}
}

func StartScheduler() {
	logger.Info("Scheduler is starting...")
	muStop.Lock()
	isStop = false
	stopCond.Broadcast()
	muStop.Unlock()
}

func StopScheduler() {
	logger.Info("Scheduler is stopping...")
	muStop.Lock()
	isStop = true
	stopCond.Broadcast()
	muStop.Unlock()
}

func GetTask() (string, error) {
	muStop.Lock()
	for isStop {
		stopCond.Wait()
	}
	muStop.Unlock()

	muTasks.Lock()
	defer muTasks.Unlock()

	if len(manualTasks) > 0 {
		task := manualTasks[0]
		manualTasks = manualTasks[1:]
		return task, nil
	}

	triggerFetch := func() {
		if !fetchInProgress {
			fetchInProgress = true
			go func() {
				logger.Infof("Pending tasks less than %d, fetching tasks from database", FetchThreshold)
				err := fetchTasksFromDatabase()
				muTasks.Lock()
				fetchInProgress = false
				muTasks.Unlock()
				taskCond.Broadcast() // Notify waiting processes
				if err != nil {
					logger.Errorf("Error fetching tasks: %v\n", err)
				}
			}()
		}

	}

	// Trigger fetch if tasks are below FetchThreshold
	if len(tasks) < FetchThreshold {
		triggerFetch()
	}

	// Wait if tasks are empty and fetch is in progress
	for len(tasks) == 0 {
		triggerFetch()
		taskCond.Wait()
	}

	if len(tasks) > 0 {
		task := tasks[0]
		tasks = tasks[1:]
		return task, nil
	}

	return "", fmt.Errorf("no task available")
}

// In order to unique incoming tasks (when get data from database, there exists
// tasks which is same with executing one), so executing tasks are store in a set,
// when task is finished, this method is supposed to be called.
func FinishTask(t string) {
	muToDelete.Lock()
	defer muToDelete.Unlock()
	toDeleteTasks = append(toDeleteTasks, t)
}

func GetPendingTasks() []string {
	muTasks.Lock()
	defer muTasks.Unlock()

	t := slices.Clone(manualTasks)
	t = append(t, tasks...)
	return t
}

func IsScheduleRunning() bool {
	muStop.Lock()
	defer muStop.Unlock()
	return !isStop
}
