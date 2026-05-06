package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type UserSource string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusLocked   UserStatus = "locked"

	UserSourceLocal UserSource = "local"
	UserSourceLDAP  UserSource = "ldap"
	UserSourceADFS  UserSource = "adfs"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Username     string     `db:"username"`
	Email        string     `db:"email"`
	DisplayName  string     `db:"display_name"`
	Status       UserStatus `db:"status"`
	Source       UserSource `db:"source"`
	PasswordHash string     `db:"password_hash"`
	FailedLogins int        `db:"failed_logins"`
	LockedUntil  *time.Time `db:"locked_until"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

func (u *User) IsActive() bool {
	if u.Status != UserStatusActive {
		return false
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return false
	}
	return true
}

type TOTPSecret struct {
	ID         uuid.UUID  `db:"id"`
	UserID     uuid.UUID  `db:"user_id"`
	SecretEnc  []byte     `db:"secret_enc"` // AES-256-GCM encrypted
	Enrolled   bool       `db:"enrolled"`
	EnrolledAt *time.Time `db:"enrolled_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}
