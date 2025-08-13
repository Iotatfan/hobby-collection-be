package text

const (
	ErrInvLogin        string = "invalid email or password"
	ErrReqBody         string = "invalid request body"
	ErrBadReq          string = "bad request"
	ErrServer          string = "internal server error"
	ErrNoRegis         string = "user has not registered yet"
	ErrInvGoogle       string = "invalid google token"
	NoAuth             string = "not authorized"
	InvToken           string = "token is invalid or expired"
	CategoryExist      string = "category name already exist"
	CategoryUsageExist string = "category is used, cannot be deleted"
	CategoryNotExist   string = "category name does not exist"
	CategoryMustBeNum  string = "category_id must be a number"
)
