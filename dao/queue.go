package dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type QueueState string

const (
	Pending   = "pending"
	Executing = "executing"
	Done      = "done"
	Fail      = "fail"
)

type QueueName string

type QueueTask struct {
	Queue QueueName
	Mark  string
	Value interface{}
	State QueueState
}

func NewQueueTask(queue QueueName, mark string, value interface{}) (task QueueTask) {
	return QueueTask{
		Queue: queue,
		Mark:  mark,
		Value: value,
		State: Pending,
	}
}

type QueueTaskList []QueueTask

func markTaskAs(tx *sqlx.Tx, task QueueTask) (err error) {
	query := `update %s set state = $1 where %s = $2;`
	query = fmt.Sprintf(query, task.Queue, task.Mark)

	_, err = tx.Exec(query, task.State, task.Value)

	return
}

func pushTaskToQueue(tx *sqlx.Tx, task QueueTask) (err error) {
	query := `insert into %s (%s, state) values ($1, 'pending') on conflict (%s) do nothing;`
	query = fmt.Sprintf(query, task.Queue, task.Mark, task.Mark)

	_, err = tx.Exec(query, task.Value)

	return
}

func popTaskFromQueue(tx *sqlx.Tx, queueName QueueName, taskMark string, limit uint) (list QueueTaskList, err error) {
	sel := `select %s from %s where state = 'pending' order by added limit $1;`
	sel = fmt.Sprintf(sel, taskMark, queueName)

	var rows *sqlx.Rows
	rows, err = tx.Queryx(sel, limit)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		task = QueueTask{
			Queue: queueName,
			Mark:  taskMark,
			Value: nil,
			State: Executing,
		}
		ids []string
	)
	for rows.Next() {
		err = rows.Scan(&task.Value)
		if err != nil {
			return
		}

		list = append(list, task)

		ids = append(ids, task.Value.(string))
	}

	if len(ids) == 0 {
		return
	}

	upd := `update %s set state = 'executing', updated = now() where %s in (%s);`
	upd = fmt.Sprintf(upd, queueName, taskMark, "'"+strings.Join(ids, "','")+"'")

	_, err = tx.Exec(upd)
	if err != nil {
		return
	}

	return
}

func defrostTasks(tx *sqlx.Tx, queue QueueName, duration uint32) (err error) {
	query := `update %s set state = 'pending' 
					where state = 'executing' and updated + $1 * interval '1 second' < now();`

	query = fmt.Sprintf(query, queue)

	_, err = tx.Exec(query, duration)

	return
}
