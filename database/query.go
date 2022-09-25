package database

import "github.com/go-pg/pg/orm"

type SelectQueryOptions struct {
	Limit          int
	Offset         int
	OrderBy        string
	OrderDirection string
}

func (s *SelectQueryOptions) Apply(q *orm.Query) *orm.Query {
	if s.Limit > 0 {
		q = q.Limit(s.Limit)
	}

	if s.Offset > 0 {
		q = q.Offset(s.Offset)
	}

	if s.OrderBy != "" {
		if s.OrderDirection == "" {
			s.OrderDirection = "ASC"
		}

		q = q.Order(s.OrderBy + " " + s.OrderDirection)
	}

	return q
}
