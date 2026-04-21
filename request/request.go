package request

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/udistrital/utils_oas/xray"
)

var ErrResponseDecode = errors.New("response body could not be decoded into target")

// GetWithContext makes a GET request to the given URL using the provided context.
// Checks for non-2xx HTTP status codes, and decodes the response body into target.
func GetWithContext(ctx context.Context, urlp string, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlp, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, application/xml, text/xml")

	seg := xray.BeginSegmentSec(req)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") {
		if err := xml.NewDecoder(resp.Body).Decode(target); err != nil {
			return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
		}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
		}
	}

	return resp.StatusCode, nil
}

// PostWithContext makes a POST request to the given URL using the provided context.
// It encodes body as JSON, checks for non-2xx HTTP status codes, and decodes the
// JSON response body into target.
func PostWithContext(ctx context.Context, urlp string, body, target any) (int, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return 0, fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlp, b)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	seg := xray.BeginSegmentSec(req)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return resp.StatusCode, nil
}

// PutWithContext makes a PUT request to the given URL using the provided context.
// It encodes body as JSON, checks for non-2xx HTTP status codes, and decodes the
// JSON response body into target.
func PutWithContext(ctx context.Context, urlp string, body, target any) (int, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return 0, fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlp, b)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	seg := xray.BeginSegmentSec(req)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return resp.StatusCode, nil
}

// PatchWithContext makes a PATCH request to the given URL using the provided context.
// It encodes body as JSON, checks for non-2xx HTTP status codes, and decodes the
// JSON response body into target.
func PatchWithContext(ctx context.Context, urlp string, body, target any) (int, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return 0, fmt.Errorf("could not encode request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, urlp, b)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	seg := xray.BeginSegmentSec(req)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return resp.StatusCode, nil
}

// DeleteWithContext makes a DELETE request to the given URL using the provided context.
// Checks for non-2xx HTTP status codes, and decodes the JSON response body into target.
func DeleteWithContext(ctx context.Context, urlp string, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlp, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	seg := xray.BeginSegmentSec(req)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	xray.UpdateSegment(resp, err, seg)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("%w: %w", ErrResponseDecode, err)
	}

	return resp.StatusCode, nil
}
