package radar

import (
	"context"
	"database/sql"
	"log"
	"strconv"
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
type RadarItem struct {
	ID    int64
	URL   string
	Title string
}

type RadarItemsService struct {
	// Database to use as backend.
	Database *sql.DB
}

// List returns a list of all radar items.
func (rs RadarItemsService) List(ctx context.Context, limit int) ([]RadarItem, error) {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if limit < 0 {
		limit = 1000
	}

	rows, err := tx.Query("SELECT id, url, title FROM radar_items LIMIT 0,?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RadarItem{}
	for rows.Next() {
		var item RadarItem
		if err := rows.Scan(&item.ID, &item.URL, &item.Title); err != nil {
			return nil, err
		}
		log.Printf("loaded row=%#v", item)
		items = append(items, item)
	}

	if err = tx.Commit(); err != nil {
		return items, err
	}

	return items, nil
}

// Create adds a RadarItem to the database.
func (rs RadarItemsService) Create(ctx context.Context, m RadarItem) error {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO radar_items (urls, titles) VALUES (?, ?)")
	if err != nil {
		return err
	}

	if _, err = tx.Exec(m.URL, m.Title); err != nil {
		return err
	}
	defer stmt.Close()

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a RadarItem from the database by its ID.
func (rs RadarItemsService) Delete(ctx context.Context, id int64) error {
	tx, err := rs.Database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM radar_items WHERE id = ?")
	if err != nil {
		return err
	}
	if _, err = tx.Exec(strconv.FormatInt(id, 10)); err != nil {
		return err
	}
	defer stmt.Close()

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Shutdown closes the database connection.
func (rs RadarItemsService) Shutdown(ctx context.Context) {
	rs.Database.Close()
}
