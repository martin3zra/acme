package app

func (s *Server) findUserByEmail(email string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where email = $1 ", email).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword)
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
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) updateProfile(uuid string, form *StoreProfileForm) error {
	_, err := s.db.Exec(`
    UPDATE users
    SET name = $2, email = $3
    WHERE id = (SELECT owner_id FROM accounts WHERE uuid = $1)
  `, uuid, form.Name, form.Email)
	return err
}

func (s *Server) findUserByAccountUUID(uuid string) (*User, error) {
	user := new(User)

	err := s.db.QueryRow("SELECT * FROM users where id = (SELECT owner_id FROM accounts WHERE uuid = $1) ", uuid).
		Scan(&user.Id, &user.Name, &user.Email,
			&user.Password, &user.EmailVerifiedAt, &user.LastPasswordReset, &user.CreatedAt,
			&user.UpdatedAt, &user.DeletedAt, &user.UUID, &user.Status, &user.MustChangePassword)
	if err != nil {
		return nil, err
	}
	return user, nil
}
