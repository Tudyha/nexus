package middleware

func Init() error {
	if err := initAuthMiddleware(); err != nil {
		return err
	}
	return nil
}
