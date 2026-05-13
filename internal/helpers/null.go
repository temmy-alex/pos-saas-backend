package helpers

import (
	"database/sql"
	"time"
)

func NullableString(value sql.NullString) *string {
	if value.Valid {
		return &value.String
	}

	return nil
}

func NullableInt64(value sql.NullInt64) *int64 {
	if value.Valid {
		return &value.Int64
	}

	return nil
}

func NullableTime(value sql.NullTime) *time.Time {
	if value.Valid {
		return &value.Time
	}

	return nil
}
