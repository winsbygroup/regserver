package activation

type Request struct {
	MachineCode string `json:"machineCode"`
	UserName    string `json:"userName"`
}

type Response struct {
	UserName            string         `json:"UserName"`
	UserCompany         string         `json:"UserCompany"`
	MachineCode         string         `json:"MachineCode"`
	ExpirationDate      string         `json:"ExpirationDate"`
	MaintExpirationDate string         `json:"MaintExpirationDate"`
	MaxProductVersion   string         `json:"MaxProductVersion"`
	LatestVersion       string         `json:"LatestVersion"`
	ProductGUID         string         `json:"ProductGUID"`
	LicenseKey          string         `json:"LicenseKey"`
	RegistrationHash    string         `json:"RegistrationHash"`
	Features            map[string]any `json:"Features"`
}
