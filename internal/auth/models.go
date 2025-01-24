package auth

type User struct {
	ID       string
	Username string
	Password string
	Role     string
}

type Token struct {
	UserID string
	Token  string
}
