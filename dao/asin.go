package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/call2mall/conn"
	"github.com/jackc/pgx/v4"
	"path/filepath"
)

type ASIN string

func (a ASIN) FilePath(baseDir string) (filePath string) {
	prefix := a[0:4]
	filePath = filepath.Join(baseDir, fmt.Sprintf("%s/%s.jpg", prefix, a))

	return
}

func (a ASIN) MarkGrabberAs(state QueueState) (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		task := NewQueueTask("asin.grabber_queue", "asin", a)
		task.State = state

		err = markTaskAs(tx, task)

		return
	})

	return
}

type ASINList []ASIN

func (l ASINList) Store() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `insert into asin.list (asin) values ($1) on conflict (asin) do nothing;`

		for _, asin := range l {
			_, err = tx.Exec(context.Background(), query, asin)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func (l ASINList) PushToGrabberQueue() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		var task QueueTask

		for _, asin := range l {
			task = NewQueueTask("asin.grabber_queue", "asin", asin)

			err = pushTaskToQueue(tx, task)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func PopFromGrabberQueue(limit uint) (list ASINList, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		var taskList QueueTaskList
		taskList, err = popTaskFromQueue(tx, "asin.grabber_queue", "asin", limit)
		if err != nil {
			return
		}

		for _, t := range taskList {
			list = append(list, ASIN(t.Value.(string)))
		}

		return
	})

	return
}

func DefrostGrabberQueue(duration uint32) (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		err = defrostTasks(tx, "asin.grabber_queue", duration)

		return
	})

	return
}

func (l ASINList) Diff(o ASINList) (d ASINList) {
	var heap = map[ASIN]interface{}{}
	for _, a := range l {
		heap[a] = nil
	}

	var ok bool
	for _, a := range o {
		_, ok = heap[a]
		if !ok {
			heap[a] = nil

			d = append(d, a)
		}
	}

	return
}

type ASINProps struct {
	ASIN ASIN

	Image Image

	ASINMeta
}

type ASINMeta struct {
	Url       string
	Category  Category
	Title     string
	L8n       string
	ImageHash string
}

func GetProps(asin ASIN) (props ASINProps, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `select c.id, c.name, l.url, l.title, l.l8n, i.bytes, i.hash 
					from asin.list l 
					join asin.image i on l.image_hash = i.hash
					join asin.category c on l.category_id = c.id
					where l.asin = $1;`

		var (
			category Category
			image    Image

			asinL8n sql.NullString
		)
		err = tx.QueryRow(context.Background(), query, asin).Scan(&category.Id, &category.Name,
			&props.Url, &props.Title, &asinL8n, &image.Bytes, &props.ImageHash)
		if err == pgx.ErrNoRows {
			err = nil

			return
		}

		if asinL8n.Valid {
			props.Title = asinL8n.String
		}

		props.ASIN = asin
		props.Category = category
		props.Image = image

		return
	})

	return
}

func (props ASINProps) Store() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		var categoryId uint32
		categoryId, err = props.Category.store(tx)
		if err != nil {
			return
		}

		err = props.Image.store(tx)
		if err != nil {
			return
		}

		upd := `update asin.list set url = $2, category_id = $3, title = $4, l8n = $5, image_hash = $6 where asin = $1;`

		_, err = tx.Exec(context.Background(), upd, props.ASIN, props.Url, categoryId, sql.NullString{
			String: props.Title,
			Valid:  len(props.Title) > 0,
		}, sql.NullString{
			String: props.L8n,
			Valid:  len(props.L8n) > 0,
		}, props.Image.Hash())

		return
	})

	return
}

func GetPublishedASIN() (al ASINList, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `select u.asin from catalog.unit u where u.is_published and not u.is_remove;`

		var rows pgx.Rows
		rows, err = tx.Query(context.Background(), query)
		if err != nil {
			return
		}
		defer rows.Close()

		var a ASIN
		for rows.Next() {
			err = rows.Scan(&a)
			if err != nil {
				return
			}

			al = append(al, a)
		}

		return
	})

	return
}

func GetAllASIN() (al ASINList, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `select l.asin from asin.list l;`

		var rows pgx.Rows
		rows, err = tx.Query(context.Background(), query)
		if err != nil {
			return
		}
		defer rows.Close()

		var a ASIN
		for rows.Next() {
			err = rows.Scan(&a)
			if err != nil {
				return
			}

			al = append(al, a)
		}

		return
	})

	return
}

func (a ASIN) Publish() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `update catalog.unit set is_published = true where asin = $1;`
		_, err = tx.Exec(context.Background(), query, a)

		return
	})

	return
}
