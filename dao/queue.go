package dao

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
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

func markTaskAs(tx pgx.Tx, task QueueTask) (err error) {
	query := `update %s set state = $1 where %s = $2;`
	query = fmt.Sprintf(query, task.Queue, task.Mark)

	_, err = tx.Exec(context.Background(), query, task.State, task.Value)

	return
}

func pushTaskToQueue(tx pgx.Tx, task QueueTask) (err error) {
	query := `insert into %s (%s, state) values ($1, 'pending') on conflict (%s) do nothing;`
	query = fmt.Sprintf(query, task.Queue, task.Mark, task.Mark)

	_, err = tx.Exec(context.Background(), query, task.Value)

	return
}

func popTaskFromQueue(tx pgx.Tx, queueName QueueName, taskMark string, limit uint) (list QueueTaskList, err error) {
	sel := `select %s from %s where state = 'pending' order by added limit $1;`
	sel = fmt.Sprintf(sel, taskMark, queueName)

	var rows pgx.Rows
	rows, err = tx.Query(context.Background(), sel, limit)
	if err != nil {
		return
	}
	defer func() {
		rows.Close()
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

	_, err = tx.Exec(context.Background(), upd)
	if err != nil {
		return
	}

	return
}

func defrostTasks(tx pgx.Tx, queue QueueName, duration uint32) (err error) {
	query := `update %s set state = 'pending' 
					where state = 'executing' and updated + $1 * interval '1 second' < now();`

	query = fmt.Sprintf(query, queue)

	_, err = tx.Exec(context.Background(), query, duration)

	return
}
