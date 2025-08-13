package helper

type DBError struct {
	ErrorMsg error
}

func (db DBError) Error() string {
	return db.ErrorMsg.Error()
}

type ValError struct {
	ErrorMsg error
}

func (ve ValError) Error() string {
	return ve.ErrorMsg.Error()
}

type NoRouteError struct {
	ErrorMsg string
}

func (ne NoRouteError) Error() string {
	return ne.ErrorMsg
}

type BcryptError struct {
	ErrorMsg error
}

func (be BcryptError) Error() string {
	return be.ErrorMsg.Error()
}

type JWTError struct {
	ErrorMsg string
}

func (je JWTError) Error() string {
	return je.ErrorMsg
}

type ServiceError struct {
	ErrorMsg string
	Code     int
}

func (se ServiceError) Error() string {
	return se.ErrorMsg
}
