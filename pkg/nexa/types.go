package nexa

// General Protocol Constants
const (
	// Standard Ports
	PORT_SERVER = "1413"
	PORT_DNS    = "1112"
	PORT_WEB    = "8080"

	// Protocol Commands
	CMD_PING    = "PING"
	CMD_FETCH   = "FETCH"
	CMD_PUBLISH = "PUBLISH"
	CMD_LIST    = "LIST"
	CMD_AUTH    = "AUTH"

	// DNS Commands
	DNS_PING     = "PING"
	DNS_RESOLVE  = "RESOLVE"
	DNS_REGISTER = "REGISTER"
	DNS_UPDATE   = "UPDATE"
	DNS_DELETE   = "DELETE"
	DNS_LIST     = "LIST"

	// Response Codes
	STATUS_OK           = 200
	STATUS_CREATED      = 201
	STATUS_BAD_REQ      = 400
	STATUS_UNAUTHORIZED = 401
	STATUS_FORBIDDEN    = 403
	STATUS_NOT_FOUND    = 404
	STATUS_CONFLICT     = 409
	STATUS_SERVER_ERROR = 500
)

// Request defines the structure of a Nexa protocol request
type Request struct {
	Command string `json:"command"`
	Target  string `json:"target,omitempty"`
	Body    string `json:"body,omitempty"`
	Token   string `json:"token,omitempty"` // For auth
}

// Response defines the structure of a Nexa protocol response
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Body    string `json:"body,omitempty"`
}

// DNSRecord defines the structure of a DNS entry
type DNSRecord struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Service   string `json:"service"`
	Owner     string `json:"owner,omitempty"` // User who owns this record
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
