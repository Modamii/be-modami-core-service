package handler

// StandardSuccessEnvelope documents pkg-gokit/response OK/Created JSON shape (success + data + optional meta).
type StandardSuccessEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// StandardErrorEnvelope documents pkg-gokit/response error JSON shape.
type StandardErrorEnvelope struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}
