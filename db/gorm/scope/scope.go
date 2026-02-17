package scope

import "gorm.io/gorm"

type ScopeFunc func(db *gorm.DB) *gorm.DB
