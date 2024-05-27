package helper

type StringError string

func (s StringError) Error() string {
    return string(s)
}

func (s StringError) String() string {
    return string(s)
}

func NewError(s string) error {
    return StringError(s)
}
