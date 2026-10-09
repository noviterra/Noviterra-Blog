package interceptor

import "Noviterra/server/utils/ctxdata"

type Interceptor struct {
	CtxData *ctxdata.CtxData
}

func NewInterceptor(ctxdata *ctxdata.CtxData) *Interceptor {
	return &Interceptor{
		CtxData: ctxdata,
	}
}
