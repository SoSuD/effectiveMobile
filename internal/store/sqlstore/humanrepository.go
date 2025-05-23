package sqlstore

import (
	"context"
	"effectiveMobile/internal/model"
	"effectiveMobile/internal/store"
	"fmt"
	"strings"
)

type HumanRepository struct {
	store *Store
}

func (h *HumanRepository) AddHuman(ctx context.Context, human *model.Human) error {
	query := `INSERT INTO people (name, surname, patronymic, age, gender, nationality) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := h.store.db.Exec(ctx, query, human.Name, human.Surname, human.Patronymic, human.Age, human.Gender, human.Nationality)
	if err != nil {
		return err
	}
	return nil
}

func (h *HumanRepository) DeleteHuman(ctx context.Context, id int) error {
	const query = `
        DELETE FROM people
         WHERE id = $1
    `
	tag, err := h.store.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrHumanNotFound
	}
	return nil
}

func (h *HumanRepository) UpdateHuman(ctx context.Context, human *model.Human) error {
	const query = `
        UPDATE people
           SET name        = $1,
               surname     = $2,
               patronymic  = $3,
               age         = $4,
               gender      = $5,
               nationality = $6
         WHERE id = $7
    `
	tag, err := h.store.db.Exec(ctx, query,
		human.Name,
		human.Surname,
		human.Patronymic,
		human.Age,
		human.Gender,
		human.Nationality,
		human.Id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrHumanNotFound
	}
	return nil
}

func (h *HumanRepository) GetHumans(ctx context.Context, f *model.HumanFilter) ([]model.Human, error) {
	var sb strings.Builder
	sb.WriteString(`
        SELECT
            id, name, surname, patronymic,
            age, gender, nationality
        FROM people
    `)

	var whereClauses []string
	var args []interface{}

	if f.Name != "" {
		args = append(args, "%"+f.Name+"%")
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if f.Surname != "" {
		args = append(args, "%"+f.Surname+"%")
		whereClauses = append(whereClauses, fmt.Sprintf("surname ILIKE $%d", len(args)))
	}
	if f.Patronymic != "" {
		args = append(args, "%"+f.Patronymic+"%")
		whereClauses = append(whereClauses, fmt.Sprintf("patronymic ILIKE $%d", len(args)))
	}
	if f.MinAge > 0 {
		args = append(args, f.MinAge)
		whereClauses = append(whereClauses, fmt.Sprintf("age >= $%d", len(args)))
	}
	if f.MaxAge > 0 {
		args = append(args, f.MaxAge)
		whereClauses = append(whereClauses, fmt.Sprintf("age <= $%d", len(args)))
	}
	if f.Gender != "" {
		args = append(args, f.Gender)
		whereClauses = append(whereClauses, fmt.Sprintf("gender = $%d", len(args)))
	}
	if f.Nationality != "" {
		args = append(args, f.Nationality)
		whereClauses = append(whereClauses, fmt.Sprintf("nationality = $%d", len(args)))
	}
	if f.ID > 0 {
		args = append(args, f.ID)
		whereClauses = append(whereClauses, fmt.Sprintf("id = $%d", len(args)))
	}

	if len(whereClauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(whereClauses, " AND "))
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize

	args = append(args, f.PageSize, offset)
	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)))

	query := sb.String()

	rows, err := h.store.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var humans []model.Human
	for rows.Next() {
		var h model.Human
		if err := rows.Scan(
			&h.Id,
			&h.Name,
			&h.Surname,
			&h.Patronymic,
			&h.Age,
			&h.Gender,
			&h.Nationality,
		); err != nil {
			return nil, err
		}
		humans = append(humans, h)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return humans, nil
}
