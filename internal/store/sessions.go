package store

import "time"

func (s *Store) CreateSession(tokenHash string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (token_hash, expires_at) VALUES (?, ?)`,
		tokenHash, expiresAt.Unix(),
	)
	return err
}

func (s *Store) SessionValid(tokenHash string, now time.Time) (bool, error) {
	var valid int
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM sessions WHERE token_hash = ? AND expires_at > ?)`,
		tokenHash, now.Unix(),
	).Scan(&valid)
	if err != nil {
		return false, err
	}
	return valid == 1, nil
}

func (s *Store) DeleteSession(tokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

func (s *Store) DeleteAllSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions`)
	return err
}

func (s *Store) DeleteSessionsExcept(tokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash <> ?`, tokenHash)
	return err
}

func (s *Store) DeleteExpiredSessions(now time.Time) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, now.Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
