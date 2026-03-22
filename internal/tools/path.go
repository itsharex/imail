// Copyright 2018 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tools

import (
	"net/url"
	"path/filepath"
	"strings"
)

// IsSameSiteURLPath returns true if the URL path belongs to the same site, false otherwise.
// False: //url, http://url, /\url, @url, ://url
// True: /url
func IsSameSiteURLPath(path string) bool {
	// Must start with / and not have protocol or host indicators
	if len(path) < 2 || path[0] != '/' {
		return false
	}
	// Check for second character being / or \ (protocol-relative or path escape)
	if path[1] == '/' || path[1] == '\\' {
		return false
	}
	// Check for URL-encoded characters that could be used for bypass
	if strings.Contains(path, "%") {
		decoded, err := url.QueryUnescape(path)
		if err != nil {
			return false
		}
		// Re-check after decoding
		if len(decoded) >= 2 && (decoded[1] == '/' || decoded[1] == '\\') {
			return false
		}
		// Check for @ which could indicate userinfo in URL
		if strings.Contains(decoded, "@") {
			return false
		}
	}
	// Check for @ which could indicate userinfo in URL (e.g., /@example.com)
	if strings.Contains(path, "@") {
		return false
	}
	// Check for :// which indicates a scheme
	if strings.Contains(path, "://") {
		return false
	}
	return true
}

// IsMaliciousPath returns true if given path is an absolute path or contains malicious content
// which has potential to traverse upper level directories.
func IsMaliciousPath(path string) bool {
	return filepath.IsAbs(path) || strings.Contains(path, "..")
}
