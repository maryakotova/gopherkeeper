package customererrors

var ErrUsernameTaken = &MyError{Message: "логин уже занят"}

type MyError struct {
	Message string
}

func (e *MyError) Error() string {
	return e.Message
}
