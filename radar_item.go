package radar

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// RadarItem is a single row in the radar_items table. It contains a URL and optionally a title.
//
// The table is defined thusly:
// CREATE TABLE `radar_items` (
//   `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
//   `url` text NOT NULL,
//   `title` text,
//   PRIMARY KEY (`id`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
//
//
// RadarItem.GetTitle() is defined in parser.go. Use that to fetch the title!
type RadarItem struct {
	ID    int64
	URL   string
	Title string

	parsedURL *url.URL
}

func (r *RadarItem) GetHostname() string {
	if r.parsedURL == nil {
		var err error
		r.parsedURL, err = url.Parse(r.URL)
		if err != nil {
			log.Printf("GetHostname: couldn't parse URL %q: %+v", r.URL, err)
			return ""
		}
	}

	return r.parsedURL.Hostname()
}

type RadarItems []RadarItem

func (r RadarItems) Len() int {
	return len(r)
}

func (r RadarItems) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RadarItems) Less(i, j int) bool {
	return r[i].GetHostname() < r[j].GetHostname()
}

type RadarItemsService struct {
	// Database to use as backend.
	Database *sql.DB
}

// List returns a list of all radar items.
func (rs RadarItemsService) List(ctx context.Context, limit int) ([]RadarItem, error) {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed to begin")
	}
	defer tx.Rollback()

	if limit < 0 {
		limit = 1000
	}

	rows, err := tx.Query("SELECT id, url, title FROM radar_items LIMIT 0,?", limit)
	if err != nil {
		return nil, errors.Wrap(err, "query for select failed")
	}
	defer rows.Close()

	items := []RadarItem{}
	for rows.Next() {
		var item RadarItem
		var title sql.NullString
		if err := rows.Scan(&item.ID, &item.URL, &title); err != nil {
			return nil, errors.Wrap(err, "scan for select failed")
		}
		log.Printf("loaded row=%#v", item)
		if title.Valid {
			item.Title = title.String
		}
		items = append(items, item)
	}

	if err = tx.Commit(); err != nil {
		return items, errors.Wrap(err, "commit for select failed")
	}

	return items, nil
}

// Delete removes a RadarItem from the database by its ID.
func (rs RadarItemsService) Get(ctx context.Context, id int64) (RadarItem, error) {
	var radarItem RadarItem

	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return radarItem, errors.Wrap(err, "transaction failed to begin")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("SELECT id, url, title FROM radar_items WHERE id = ?")
	if err != nil {
		return radarItem, errors.Wrap(err, "prepare for get failed")
	}

	if err = stmt.QueryRow(strconv.FormatInt(id, 10)).Scan(&radarItem.ID, &radarItem.URL, &radarItem.Title); err != nil {
		return radarItem, errors.Wrap(err, "queryrow for get failed")
	}
	defer stmt.Close()

	err = tx.Commit()
	if err != nil {
		return radarItem, errors.Wrap(err, "commit for get failed")
	}

	return radarItem, nil
}

// Create adds a RadarItem to the database.
func (rs RadarItemsService) Create(ctx context.Context, m RadarItem) error {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "transaction failed to begin")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO radar_items (url, title) VALUES ( ?, ? )")
	if err != nil {
		return errors.Wrap(err, "prepare for insert failed")
	}

	if _, err = stmt.Exec(m.URL, m.Title); err != nil {
		return errors.Wrap(err, "exec for insert failed")
	}
	defer stmt.Close()

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit for insert failed")
	}

	return nil
}

// Delete removes a RadarItem from the database by its ID.
func (rs RadarItemsService) Delete(ctx context.Context, id int64) error {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "transaction failed to begin")
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM radar_items WHERE id = ?")
	if err != nil {
		return errors.Wrap(err, "prepare for delete failed")
	}
	if _, err = stmt.Exec(strconv.FormatInt(id, 10)); err != nil {
		return errors.Wrap(err, "exec for delete failed")
	}
	defer stmt.Close()

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit for delete failed")
	}

	return nil
}

// Shutdown closes the database connection.
func (rs RadarItemsService) Shutdown(ctx context.Context) {
	rs.Database.Close()
}
