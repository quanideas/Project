package response

type BaseResponse struct {
	Meta struct {
		Status int
	}
	Data interface{}
}

type GetAll struct {
	List  interface{}
	Total int64
}
