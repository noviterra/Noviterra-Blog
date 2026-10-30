package errcode

/*
封装repo中返回的err
区分原生err和自定义err
*/

type RepoError struct {
	Msg string
}

func NewRepoError(msg string) *RepoError {
	return &RepoError{
		Msg: msg,
	}
}

func (re *RepoError) Error() string {
	return re.Msg
}
