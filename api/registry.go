package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wercker/wercker/util"
)

// StepRegistry abstracts registry interaction
type StepRegistry interface {
	// GetStepVersion retrieves a step from the registry
	GetStepVersion(owner, name, version string) (*APIStepVersion, error)
	// GetTarball retrieves a step tarball from the registry
	GetTarball(tarballURL string) (*http.Response, error)
}

// WerckerStepRegistry implements the StepRegistry interface to handle
type WerckerStepRegistry struct {
	baseURL string
	authToken string
}

// NewWerckerStepRegistry creates a new instance of NewWerckerStepRegistry
func NewWerckerStepRegistry(baseURL, authToken string) StepRegistry {
	return &WerckerStepRegistry{
		baseURL: baseURL,
		authToken: authToken,
	}
}

// GetStepVersion retrieves a step from the registry
func (r *WerckerStepRegistry) GetStepVersion(owner, name, version string) (*APIStepVersion, error) {
	url := fmt.Sprintf("%s/api/steps/%s/%s/%s", r.baseURL, owner, name, version)

	resp, err := util.Get(url, r.authToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
		}
	}

	stepVersion := struct {
		Step struct {
			Summary    string `json:"summary"`
			TarballURL string `json:"tarballUrl"`
			Version    struct {
				Number string `json:"number"`
			} `json:"version"`
		} `json:"step"`
	}{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&stepVersion); err != nil {
		return nil, err
	}

	return &APIStepVersion{
		Description: stepVersion.Step.Summary,
		TarballURL:  stepVersion.Step.TarballURL,
		Version:     stepVersion.Step.Version.Number,
	}, nil
}

func (r *WerckerStepRegistry) GetTarball(tarballURL string) (*http.Response, error) {
	return util.Get(tarballURL, r.authToken)
}
