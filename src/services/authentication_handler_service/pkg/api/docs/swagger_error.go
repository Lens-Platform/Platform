package docs

// Error occured during request lifecycle
// swagger:response genericError
type genericErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}

// Bad Request
// swagger:response badRequestError
type badRequestErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}

// Forbidden Request
// swagger:response forbiddenError
type forbiddenErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}

// UnAuthorized Request
// swagger:response unAuthorizedError
type unAuthorizedErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}

// Request Not Found
// swagger:response notFoundError
type notFoundErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}

// Internal Server Error
// swagger:response internalServerError
type internalServerErrorResponse struct {
	// in: body
	Body struct {
		// description of the error
		// required: true
		Error error `json:"error"`
	}
}
