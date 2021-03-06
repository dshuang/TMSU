/*
Copyright 2011-2014 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package database

import (
	"database/sql"
	"strings"
	"tmsu/entities"
)

// The number of tags in the database.
func (db *Database) TagCount() (uint, error) {
	sql := `SELECT count(1)
			FROM tag`

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// The set of tags.
func (db *Database) Tags() (entities.Tags, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTags(rows, make(entities.Tags, 0, 10))
}

// Retrieves a specific tag.
func (db *Database) Tag(id entities.TagId) (*entities.Tag, error) {
	sql := `SELECT id, name
	        FROM tag
	        WHERE id = ?`

	rows, err := db.ExecQuery(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTag(rows)
}

// Retrieves a specific set of tags.
func (db *Database) TagsByIds(ids entities.TagIds) (entities.Tags, error) {
	sql := `SELECT id, name
	        FROM tag
	        WHERE id IN (?`
	sql += strings.Repeat(",?", len(ids)-1)
	sql += ")"

	params := make([]interface{}, len(ids))
	for index, id := range ids {
		params[index] = id
	}

	rows, err := db.ExecQuery(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags, err := readTags(rows, make(entities.Tags, 0, len(ids)))
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// Retrieves a specific tag.
func (db *Database) TagByName(name string) (*entities.Tag, error) {
	sql := `SELECT id, name
	        FROM tag
	        WHERE name = ?`

	rows, err := db.ExecQuery(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTag(rows)
}

// Retrieves the set of named tags.
func (db *Database) TagsByNames(names []string) (entities.Tags, error) {
	if len(names) == 0 {
		return make(entities.Tags, 0), nil
	}

	sql := `SELECT id, name
            FROM tag
            WHERE name IN (?`
	sql += strings.Repeat(",?", len(names)-1)
	sql += ")"

	params := make([]interface{}, len(names))
	for index, name := range names {
		params[index] = name
	}

	rows, err := db.ExecQuery(sql, params...)
	if err != nil {
		return nil, err
	}

	tags, err := readTags(rows, make(entities.Tags, 0, len(names)))
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// Adds a tag.
func (db *Database) InsertTag(name string) (*entities.Tag, error) {
	sql := `INSERT INTO tag (name)
	        VALUES (?)`

	result, err := db.Exec(sql, name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		panic("expected exactly one row to be affected.")
	}

	return &entities.Tag{entities.TagId(id), name}, nil
}

// Renames a tag.
func (db *Database) RenameTag(tagId entities.TagId, name string) (*entities.Tag, error) {
	sql := `UPDATE tag
	        SET name = ?
	        WHERE id = ?`

	result, err := db.Exec(sql, name, tagId)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		panic("expected exactly one row to be affected.")
	}

	return &entities.Tag{tagId, name}, nil
}

// Deletes a tag.
func (db *Database) DeleteTag(tagId entities.TagId) error {
	sql := `DELETE FROM tag
	        WHERE id = ?`

	result, err := db.Exec(sql, tagId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 1 {
		panic("expected only one row to be affected.")
	}

	return nil
}

// Retrieves the usage of each tag
func (db *Database) TagUsage() ([]entities.TagFileCount, error) {
	sql := `SELECT t.id, t.name, count(file_id)
            FROM file_tag ft, tag t
            WHERE ft.tag_id = t.id
            GROUP BY t.id
            ORDER BY t.name`

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]entities.TagFileCount, 0, 10)
	for {
		if !rows.Next() {
			break
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		var tagId entities.TagId
		var name string
		var count uint
		err := rows.Scan(&tagId, &name, &count)
		if err != nil {
			return nil, err
		}

		tags = append(tags, entities.TagFileCount{tagId, name, count})
	}

	return tags, nil
}

// unexported

func readTag(rows *sql.Rows) (*entities.Tag, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var id entities.TagId
	var name string
	err := rows.Scan(&id, &name)
	if err != nil {
		return nil, err
	}

	return &entities.Tag{id, name}, nil
}

func readTags(rows *sql.Rows, tags entities.Tags) (entities.Tags, error) {
	for {
		tag, err := readTag(rows)
		if err != nil {
			return nil, err
		}
		if tag == nil {
			break
		}

		tags = append(tags, tag)
	}

	return tags, nil
}
