package config

import "github.com/project-hub/internal/domain"

// Manager describes configuration persistence behavior.
type Manager interface {
	Load() (domain.Config, error)
	Save(config domain.Config) error
	Exists() (bool, error)
}
