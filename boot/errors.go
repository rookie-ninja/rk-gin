package rkgin

//
//
//
//type ErrorResponse struct {
//	Code    int      `json:"code" yaml:"code"`
//	Message string   `json:"message" yaml:"message"`
//	Errors  []error  `json:"errors" yaml:"errors"`
//}
//
//func NewErrorResponse(code int, message string, errors ...error) *ErrorResponse {
//	resp := &ErrorResponse{
//		Code: code,
//		Message: message,
//		Errors: make([]error, 0),
//	}
//
//	resp.Errors = append(resp.Errors, errors...)
//
//	return resp
//}
//
//func (resp *ErrorResponse) AddError(err error) {
//	resp.Errors = append(resp.Errors, err)
//}
//
//func (resp *ErrorResponse) String() {
//
//}
//
//func (resp *ErrorResponse) JSON() {
//
//}
//
//func (resp *ErrorResponse) MarshalJSON() string {
//
//}
//
//func (resp *ErrorResponse) UnmarshalJSON() error {
//	return nil
//}
