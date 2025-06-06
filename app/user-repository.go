package app

func (s *Server) findUserByEmail(email string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where email = $1 ", email).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) findUserByUUID(uuid string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where uuid = $1 ", uuid).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) updateProfile(uuid string, form *StoreProfileForm) error {
	_, err := s.db.Exec(`
    UPDATE users
    SET name = $2, pending_email = CASE WHEN LOWER(email) <> LOWER($3) AND pending_email IS NULL THEN $3 ELSE pending_email END
    WHERE id = (SELECT owner_id FROM accounts WHERE uuid = $1)
  `, uuid, form.Name, form.Email)
	return err
}

func (s *Server) findUserByAccountUUID(uuid string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where id = (SELECT owner_id FROM accounts WHERE uuid = $1) ", uuid).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword, &user.PendingEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) updatePendingEmail(user *User) error {
	_, err := s.db.Exec("UPDATE users SET email = pending_email, pending_email = NULL WHERE id = $1 AND pending_email = $2", user.Id, user.PendingEmail)
	return err
}
