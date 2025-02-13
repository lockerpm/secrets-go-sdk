package types

type Secret struct {
	Object          string   `json:"object" `
	ID              string   `json:"id" `
	CreationDate    float64  `json:"creation_date" `
	RevisionDate    float64  `json:"revision_date" `
	UpdatedDate     *float64 `json:"updated_date" `
	DeletedDate     *float64 `json:"deleted_date" `
	LastUseDate     *float64 `json:"last_use_date" `
	ProjectID       int      `json:"project_id" `
	EnvironmentID   *string  `json:"environment_id" `
	EnvironmentName *string  `json:"environment_name" `
	EnvironmentHash *string  `json:"environment_hash" gorm:"uniqueIndex:env_sec_hash_tuple"`
	Key             string   `json:"key" `
	SecretHash      string   `json:"secret_hash" gorm:"uniqueIndex:env_sec_hash_tuple"`
	Value           string   `json:"value" `
	Description     string   `json:"description" `
}

type Environment struct {
	Object       string   `json:"object"`
	ID           string   `json:"id" `
	Name         string   `json:"name"`
	Hash         string   `json:"hash" gorm:"uniqueIndex:env_hash"`
	ExternalURL  string   `json:"external_url" `
	Description  string   `json:"description" `
	CreationDate float64  `json:"creation_date" `
	RevisionDate float64  `json:"revision_date"`
	UpdatedDate  *float64 `json:"updated_date" `
	ProjectID    int      `json:"project_id" `
}
type Project struct {
	Object         string  `json:"object"`
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	CreationDate   float64 `json:"creation_date"`
	RevisionDate   float64 `json:"revision_date"`
	UpdatedDate    float64 `json:"updated_date"`
	Key            string  `json:"key"`
	PublicKey      string  `json:"public_key"`
	PrivateKey     string  `json:"private_key"`
	OrganizationID string  `json:"organization_id"`
	CreatedBy      struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"created_by"`
}

type ProfileResponse struct {
	Object  string `json:"object"`
	Profile struct {
		Object         string   `json:"object"`
		ID             string   `json:"id"`
		ClientID       string   `json:"client_id"`
		Key            string   `json:"key"`
		Activated      bool     `json:"activated"`
		Editable       bool     `json:"editable"`
		RestrictIP     []string `json:"restrict_ip"`
		CreationDate   float64  `json:"creation_date"`
		RevisionDate   float64  `json:"revision_date"`
		ExpirationDate float64  `json:"expiration_date"`
		ProjectID      int      `json:"project_id"`
		Projects       []string `json:"projects"`
	} `json:"profile"`
}

type Profile struct {
	ProjectsStr    string
	RestrictIPStr  string
	ID             string
	ClientID       string
	Key            string
	Activated      bool
	Editable       bool
	CreationDate   float64
	RevisionDate   float64
	ExpirationDate float64
	ProjectID      int
}

type GenericList struct {
	Count        int     `json:"count"`
	Next         string  `json:"next"`
	Previous     string  `json:"previous"`
	RevisionDate float64 `json:"revision_date"`
}

type SecretResponse struct {
	Count        int      `json:"count"`
	Next         string   `json:"next"`
	Previous     string   `json:"previous"`
	RevisionDate float64  `json:"revision_date"`
	Results      []Secret `json:"results"`
}

type EnvironmentResponse struct {
	Count        int           `json:"count"`
	Next         string        `json:"next"`
	Previous     string        `json:"previous"`
	RevisionDate float64       `json:"revision_date"`
	Results      []Environment `json:"results"`
}

type EncryptedSecResponse struct {
	Object       string   `json:"object"`
	ID           string   `json:"id"`
	CreationDate float64  `json:"creation_date"`
	RevisionDate float64  `json:"revision_date"`
	UpdatedDate  *float64 `json:"updated_date"`
	DeletedDate  *float64 `json:"deleted_date"`
	LastUseDate  *float64 `json:"last_use_date"`
	Key          string   `json:"key"`
	SecretHash   string   `json:"secret_hash"`
	Value        string   `json:"value"`
	Description  string   `json:"description"`
	Data         struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"data"`
	ProjectID       int     `json:"project_id"`
	EnvironmentID   *string `json:"environment_id"`
	EnvironmentName *string `json:"environment_name"`
	EnvironmentHash *string `json:"environment_hash"`
}

type EncryptedEnvResponse struct {
	Object       string  `json:"object"`
	ID           string  `json:"id"`
	CreationDate float64 `json:"creation_date"`
	RevisionDate float64 `json:"revision_date"`
	UpdatedDate  float64 `json:"updated_date"`
	Name         string  `json:"name"`
	Hash         string  `json:"hash"`
	ExternalURL  string  `json:"external_url"`
	Description  string  `json:"description"`
	Data         struct {
		Name        string `json:"name"`
		ExternalURL string `json:"external_url"`
		Description string `json:"description"`
	} `json:"data"`
	ProjectID int `json:"project_id"`
	Project   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
}

type RevisionDate struct {
	ID           int     `gorm:"default:0"`
	RevisionDate float64 `gorm:"default:0; not null"`
	LastCallSec  float64 `gorm:"default:0; not null"`
	LastCallEnv  float64 `gorm:"default:0; not null"`
}

type DeletionDate struct {
	ID           int `gorm:"default:0"`
	DeletionDate float64
}

type ServerErrorMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DBVersion struct {
	ID               int `gorm:"default:0"`
	DbRevisionNumber int `gorm:"default:0"`
}

type ScannerInfo struct {
	ReleaseURL          string
	Binary              string
	BinaryPath          string
	BinaryPathExtracted string
}

type ResolvedInfo struct {
	Version      string
	DownloadURL  string
	ChecksumURL  string
	ArchiveName  string
	ChecksumName string
}

type GithubResponse struct {
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type TrufflehogOutput struct {
	SourceMetadata struct {
		Data struct {
			Filesystem struct {
				File string `json:"file"`
				Line int    `json:"line"`
			} `json:"Filesystem"`
		} `json:"Data"`
	} `json:"SourceMetadata"`
	Raw          string `json:"Raw"`
	DetectorName string `json:"DetectorName"`
	DecoderName  string `json:"DecoderName"`
}

type GitleaksOutput struct {
	StartLine int    `json:"StartLine"`
	Secret    string `json:"Secret"`
	File      string `json:"File"`
	Commit    string `json:"Commit"`
	RuleID    string `json:"RuleID"`
}

type NormalizedOutput struct {
	Secret     string `json:"Secret"`
	File       string `json:"File"`
	Line       int    `json:"Line"`
	SecretType string `json:"SecretType"`
	Commit     string `json:"Commit,omitempty"`
}
