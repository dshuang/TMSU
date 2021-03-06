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

package storage

import (
	"errors"
	"fmt"
	"tmsu/entities"
	"unicode"
)

// Retrievse the count of values.
func (storage *Storage) ValueCount() (uint, error) {
	return storage.Db.ValueCount()
}

// Retrieves the complete set of values.
func (storage *Storage) Values() (entities.Values, error) {
	return storage.Db.Values()
}

// Retrieves a specific value.
func (storage *Storage) Value(id entities.ValueId) (*entities.Value, error) {
	return storage.Db.Value(id)
}

// Retrieves a specific set of values.
func (storage Storage) ValuesByIds(ids entities.ValueIds) (entities.Values, error) {
	return storage.Db.ValuesByIds(ids)
}

// Retrievse the set of unused values.
func (storage *Storage) UnusedValues() (entities.Values, error) {
	return storage.Db.UnusedValues()
}

// Retrieves a specific value by name.
func (storage *Storage) ValueByName(name string) (*entities.Value, error) {
	if name == "" {
		return &entities.Value{0, ""}, nil
	}

	return storage.Db.ValueByName(name)
}

// Retrieves the set of values for the specified tag.
func (storage *Storage) ValuesByTag(tagId entities.TagId) (entities.Values, error) {
	return storage.Db.ValuesByTagId(tagId)
}

// Retrieves the set of values with the specified names.
func (storage *Storage) ValuesByNames(names []string) (entities.Values, error) {
	return storage.Db.ValuesByNames(names)
}

// Adds a value.
func (storage *Storage) AddValue(name string) (*entities.Value, error) {
	if err := validateValueName(name); err != nil {
		return nil, err
	}

	return storage.Db.InsertValue(name)
}

// Deletes a value.
func (storage *Storage) DeleteValue(valueId entities.ValueId) error {
	fileTags, err := storage.FileTagsByValueId(valueId)
	if err != nil {
		return err
	}

	for _, fileTag := range fileTags {
		if err := storage.Db.DeleteFileTag(fileTag.FileId, fileTag.TagId, fileTag.ValueId); err != nil {
			return err
		}
	}

	return storage.Db.DeleteValue(valueId)
}

// Deletes the value if it is unused.
func (storage *Storage) DeleteValueIfUnused(valueId entities.ValueId) error {
	if valueId == 0 {
		return nil
	}

	count, err := storage.FileTagCountByValueId(valueId)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.Db.DeleteValue(valueId); err != nil {
			return err
		}
	}

	return nil
}

// Deletes unused values.
func (storage *Storage) DeleteUnusedValues(valueIds entities.ValueIds) error {
	return storage.Db.DeleteUnusedValues(valueIds)
}

// unexported

var validValueChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol}

func validateValueName(valueName string) error {
	switch valueName {
	case "":
		return errors.New("tag value cannot be empty.")
	case ".", "..":
		return errors.New("tag value cannot be '.' or '..'.") // cannot be used in the VFS
	case "and", "AND", "or", "OR", "not", "NOT":
		return errors.New("tag value cannot be a logical operator: 'and', 'or' or 'not'.") // used in query language
	case "eq", "EQ", "ne", "NE", "lt", "LT", "gt", "GT", "le", "LE", "ge", "GE":
		return errors.New("tag value cannot be a comparison operator: 'eq', 'ne', 'lt', 'gt', 'le' or 'ge'.") // used in query language
	}

	for _, ch := range valueName {
		switch ch {
		case '(', ')':
			return errors.New("tag value cannot contain parentheses: '(' or ')'.") // used in query language
		case ',':
			return errors.New("tag value cannot contain comma: ','.") // reserved for tag delimiter
		case '=', '!', '<', '>':
			return errors.New("tag value cannot contain a comparison operator: '=', '!', '<' or '>'.") // reserved for tag values
		case ' ', '\t':
			return errors.New("tag value cannot contain space or tab.") // used as tag delimiter
		case '/':
			return errors.New("tag value cannot contain slash: '/'.") // cannot be used in the VFS
		}

		if !unicode.IsOneOf(validValueChars, ch) {
			return fmt.Errorf("tag value cannot contain '%c'.", ch)
		}
	}

	return nil
}
