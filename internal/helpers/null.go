package helpers

import "database/sql"

func NullableString(value sql.NullString) *string {
	if value.Valid {
		return &value.String
	}

	return nil
}
