package store

import (
	"context"
	"effectiveMobile/internal/model"
)

type HumanRepository interface {
	AddHuman(ctx context.Context, human *model.Human) error
	GetHumans(ctx context.Context, f *model.HumanFilter) ([]model.Human, error)
	UpdateHuman(ctx context.Context, human *model.Human) error
	DeleteHuman(ctx context.Context, id int) error
}
