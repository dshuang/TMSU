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
	"tmsu/entities"
)

// Retrieves the complete set of tag implications.
func (storage *Storage) Implications() (entities.Implications, error) {
	return storage.Db.Implications()
}

// Retrieves the set of implications for the specified tags.
func (storage *Storage) ImplicationsForTags(tagIds ...entities.TagId) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	impliedTagIds := make(entities.TagIds, len(tagIds))
	copy(impliedTagIds, tagIds)

	for len(impliedTagIds) > 0 {
		implications, err := storage.Db.ImplicationsForTags(impliedTagIds)
		if err != nil {
			return nil, err
		}

		impliedTagIds = make(entities.TagIds, 0)
		for _, implication := range implications {
			if !containsImplication(resultantImplications, implication) {
				resultantImplications = append(resultantImplications, implication)
				impliedTagIds = append(impliedTagIds, implication.ImpliedTag.Id)
			}
		}
	}

	return resultantImplications, nil
}

// Adds the specified implication.
func (storage Storage) AddImplication(tagId, impliedTagId entities.TagId) error {
	return storage.Db.AddImplication(tagId, impliedTagId)
}

// Updates implications featuring the specified tag.
func (storage Storage) UpdateImplicationsForTagId(tagId, impliedTagId entities.TagId) error {
	return storage.Db.UpdateImplicationsForTagId(tagId, impliedTagId)
}

// Removes the specified implication
func (storage Storage) RemoveImplication(tagId, impliedTagId entities.TagId) error {
	return storage.Db.DeleteImplication(tagId, impliedTagId)
}

// Removes implications featuring the specified tag.
func (storage Storage) RemoveImplicationsForTagId(tagId entities.TagId) error {
	return storage.Db.DeleteImplicationsForTagId(tagId)
}

// unexported

func containsImplication(implications entities.Implications, implication *entities.Implication) bool {
	for index := 0; index < len(implications); index++ {
		if implications[index].ImplyingTag.Id == implication.ImplyingTag.Id && implications[index].ImpliedTag.Id == implication.ImpliedTag.Id {
			return true
		}
	}

	return false
}
