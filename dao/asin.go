package dao

import (
	"database/sql"
	"github.com/call2mall/conn"
	"github.com/jmoiron/sqlx"
)

type ASIN string

func (a ASIN) MarkSearcherAs(state QueueState) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		task := NewQueueTask("asin.searcher_queue", "asin", a)
		task.State = state

		err = markTaskAs(tx, task)

		return
	})

	return
}

func (a ASIN) MarkEnricherAs(state QueueState) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		task := NewQueueTask("asin.enricher_queue", "asin", a)
		task.State = state

		err = markTaskAs(tx, task)

		return
	})

	return
}

func (a ASIN) MarkPublisherAs(state QueueState) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		task := NewQueueTask("asin.publisher_queue", "asin", a)
		task.State = state

		err = markTaskAs(tx, task)

		return
	})

	return
}

func (a ASIN) PushToEnricher() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		task := NewQueueTask("asin.enricher_queue", "asin", a)

		err = pushTaskToQueue(tx, task)

		return
	})

	return
}

func (a ASIN) PushToPublisher() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		task := NewQueueTask("asin.publisher_queue", "asin", a)

		err = pushTaskToQueue(tx, task)

		return
	})

	return
}

type ASINList []ASIN

func (l ASINList) Store() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `insert into asin.list (asin) values ($1) on conflict (asin) do nothing;`

		for _, asin := range l {
			_, err = tx.Exec(query, asin)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func (l ASINList) PushToSearcher() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		var task QueueTask

		for _, asin := range l {
			task = NewQueueTask("asin.searcher_queue", "asin", asin)

			err = pushTaskToQueue(tx, task)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func PopFromSearcher(limit uint) (list ASINList, err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		var taskList QueueTaskList
		taskList, err = popTaskFromQueue(tx, "asin.searcher_queue", "asin", limit)
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

func PopFromEnricher(limit uint) (list ASINList, err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		var taskList QueueTaskList
		taskList, err = popTaskFromQueue(tx, "asin.enricher_queue", "asin", limit)
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

func PopFromPublisher(limit uint) (list ASINList, err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		var taskList QueueTaskList
		taskList, err = popTaskFromQueue(tx, "asin.publisher_queue", "asin", limit)
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

func DefrostSearcher(duration uint32) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		err = defrostTasks(tx, "asin.searcher_queue", duration)

		return
	})

	return
}

func DefrostEnricher(duration uint32) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		err = defrostTasks(tx, "asin.enricher_queue", duration)

		return
	})

	return
}

func DefrostPublisher(duration uint32) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		err = defrostTasks(tx, "asin.publisher_queue", duration)

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
	Category  Category
	Title     string
	L8n       string
	ImageName string
}

func GetProps(asin ASIN) (props ASINProps, err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `select c.id, c.name, c.category_id, l.title, l.l8n, i.bytes, i.hash 
					from asin.list l 
					join asin.image i on l.image_hash = i.hash
					join asin.category c on l.category_id = c.id
					where l.asin = $1;`

		var (
			category Category
			image    Image

			asinL8n           sql.NullString
			catalogCategoryId sql.NullInt32
		)
		err = tx.QueryRowx(query, asin).Scan(&category.Id, &category.Name, &catalogCategoryId,
			&props.Title, &asinL8n, &image.Bytes, &props.ImageName)
		if err == sql.ErrNoRows {
			err = nil

			return
		}

		if catalogCategoryId.Valid {
			category.CatalogCategoryId = uint32(catalogCategoryId.Int32)
		}

		if asinL8n.Valid {
			props.Title = asinL8n.String
		}

		return
	})

	return
}

func (af ASINProps) Store() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		var categoryId uint32
		categoryId, err = af.Category.store(tx)
		if err != nil {
			return
		}

		err = af.Image.store(tx)
		if err != nil {
			return
		}

		upd := `update asin.list set category_id = $2, title = $3, l8n = $4, image_hash = $5 where asin = $1;`

		_, err = tx.Exec(upd, af.ASIN, categoryId, sql.NullString{
			String: af.Title,
			Valid:  len(af.Title) > 0,
		}, sql.NullString{
			String: af.L8n,
			Valid:  len(af.L8n) > 0,
		}, af.Image.Hash())

		return
	})

	return
}
