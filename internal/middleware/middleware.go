package middleware

func Init(jwtSecret string) error {
	if err := initAuthMiddleware(jwtSecret); err != nil {
		return err
	}
	return nil
}
